//go:build linux
// +build linux

package elevate

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
)

// runElevated on Linux uses pkexec (PolicyKit), which shows a graphical
// authentication dialog and runs the command as root with stdout/stderr
// captured normally. pkexec's own exit codes:
//
//	126 — the authorization could not be obtained (dialog dismissed/declined)
//	127 — the command was not found / could not be executed
//
// Anything else is the wrapped command's own exit code.
func runElevated(ctx context.Context, r Request) (Result, error) {
	args := append([]string{r.Command}, r.Args...)
	cmd := exec.CommandContext(ctx, "pkexec", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	if err != nil {
		code := -1
		if ee, ok := err.(*exec.ExitError); ok {
			code = ee.ExitCode()
		}
		if code == 126 {
			// User dismissed the PolicyKit dialog or wasn't authorized.
			return Result{ExitCode: code, Cancelled: true}, nil
		}
		return Result{
			OK:       false,
			ExitCode: code,
			Stdout:   strings.TrimRight(stdout.String(), "\n "),
			Stderr:   strings.TrimRight(stderr.String(), "\n "),
		}, nil
	}

	return Result{
		OK:       true,
		ExitCode: 0,
		Stdout:   strings.TrimRight(stdout.String(), "\n "),
		Stderr:   strings.TrimRight(stderr.String(), "\n "),
	}, nil
}
