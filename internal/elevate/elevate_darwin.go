//go:build darwin
// +build darwin

package elevate

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
)

// runElevated on macOS uses AppleScript's "with administrator privileges",
// which makes the OS show its native authentication dialog. The wrapped
// shell command's stdout is returned by osascript on its own stdout, so we
// capture it directly. A declined dialog makes osascript exit non-zero with
// "User canceled." on stderr (error -128).
func runElevated(ctx context.Context, r Request) (Result, error) {
	shellCmd := buildPosixCmd(r.Command, r.Args)

	// do shell script "<cmd>" with administrator privileges
	// We escape backslashes and double quotes for the AppleScript string.
	esc := strings.ReplaceAll(shellCmd, `\`, `\\`)
	esc = strings.ReplaceAll(esc, `"`, `\"`)
	script := `do shell script "` + esc + `" with administrator privileges`

	cmd := exec.CommandContext(ctx, "osascript", "-e", script)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	if err != nil {
		serr := stderr.String()
		// AppleScript reports a user cancel as error -128.
		if strings.Contains(serr, "-128") || strings.Contains(strings.ToLower(serr), "cancel") {
			return Result{ExitCode: -1, Cancelled: true}, nil
		}
		code := -1
		if ee, ok := err.(*exec.ExitError); ok {
			code = ee.ExitCode()
		}
		return Result{
			OK:       false,
			ExitCode: code,
			Stdout:   strings.TrimRight(stdout.String(), "\n "),
			Stderr:   strings.TrimRight(serr, "\n "),
		}, nil
	}

	return Result{
		OK:       true,
		ExitCode: 0,
		Stdout:   strings.TrimRight(stdout.String(), "\n "),
		Stderr:   strings.TrimRight(stderr.String(), "\n "),
	}, nil
}

// buildPosixCmd renders command + args as a POSIX shell command line,
// single-quoting tokens that need it.
func buildPosixCmd(command string, args []string) string {
	parts := append([]string{command}, args...)
	out := make([]string, len(parts))
	for i, p := range parts {
		out[i] = posixQuote(p)
	}
	return strings.Join(out, " ")
}

// posixQuote wraps a token in single quotes (the safest POSIX quoting),
// escaping any embedded single quotes via the '\'' idiom.
func posixQuote(s string) string {
	if s != "" && !strings.ContainsAny(s, " \t\"'\\$`*?&|;<>(){}[]") {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
