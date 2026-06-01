// Package runner executes external processes WITHOUT elevation, streaming
// their merged stdout/stderr to a logsink line-by-line with a timeout. Used
// for all the non-privileged probes and commands in the wizard (vboxmanage
// --version, vagrant status, vagrant ssh -c …). Privileged installs go
// through internal/elevate instead.
package runner

import (
	"bufio"
	"context"
	"io"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/abxda/vagrant-onboarding-panel/internal/logsink"
)

// Result is the outcome of a Run.
type Result struct {
	ExitCode int
	Stdout   string // captured stdout (also streamed to the sink)
	Stderr   string // captured stderr (also streamed to the sink)
	TimedOut bool
}

// Runner streams process output to a sink as it runs.
type Runner struct {
	sink *logsink.Sink
}

func New(sink *logsink.Sink) *Runner { return &Runner{sink: sink} }

// Run executes name+args with a timeout, streaming each output line to the
// sink (stdout as INFO, stderr as WARN) and also returning the captured
// text. A non-zero exit is NOT an error here — inspect Result.ExitCode.
func (r *Runner) Run(ctx context.Context, timeout time.Duration, name string, args ...string) (Result, error) {
	return r.RunDir(ctx, timeout, "", name, args...)
}

// RunDir is like Run but sets the working directory (needed for Vagrant, which
// locates its Vagrantfile via the current directory).
func (r *Runner) RunDir(ctx context.Context, timeout time.Duration, dir, name string, args ...string) (Result, error) {
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(cctx, name, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	applySysProcAttr(cmd) // platform hook (no console window on Windows)

	stdoutPipe, _ := cmd.StdoutPipe()
	stderrPipe, _ := cmd.StderrPipe()

	if r.sink != nil {
		r.sink.Emit("INFO", "$ "+quoteCmd(name, args))
	}

	if err := cmd.Start(); err != nil {
		if r.sink != nil {
			r.sink.Emit("ERROR", "No se pudo iniciar: "+err.Error())
		}
		return Result{ExitCode: -1}, err
	}

	var outBuf, errBuf strings.Builder
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); r.pump(stdoutPipe, "INFO", &outBuf) }()
	go func() { defer wg.Done(); r.pump(stderrPipe, "WARN", &errBuf) }()
	wg.Wait()

	err := cmd.Wait()
	res := Result{Stdout: outBuf.String(), Stderr: errBuf.String()}

	if cctx.Err() == context.DeadlineExceeded {
		res.TimedOut = true
		res.ExitCode = -1
		if r.sink != nil {
			r.sink.Emit("ERROR", "Tiempo de espera agotado.")
		}
		return res, nil
	}
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			res.ExitCode = ee.ExitCode()
			return res, nil
		}
		res.ExitCode = -1
		return res, err
	}
	res.ExitCode = 0
	return res, nil
}

func (r *Runner) pump(rc io.Reader, level string, buf *strings.Builder) {
	if rc == nil {
		return
	}
	sc := bufio.NewScanner(rc)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	for sc.Scan() {
		line := strings.TrimRight(sc.Text(), "\r")
		buf.WriteString(line)
		buf.WriteByte('\n')
		if r.sink != nil {
			r.sink.Emit(level, line)
		}
	}
}

func quoteCmd(name string, args []string) string {
	parts := append([]string{name}, args...)
	for i, p := range parts {
		if strings.ContainsAny(p, " \t\"") {
			parts[i] = `"` + p + `"`
		}
	}
	return strings.Join(parts, " ")
}
