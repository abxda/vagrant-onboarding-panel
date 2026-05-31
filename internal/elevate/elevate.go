// Package elevate runs single commands with elevated (administrator/root)
// privileges using each operating system's NATIVE elevation mechanism, so
// the student approves the system's own dialog at the moment the action runs.
//
//	Windows: ShellExecuteEx with the "runas" verb  → triggers UAC
//	macOS:   osascript 'do shell script … with administrator privileges'
//	Linux:   pkexec (PolicyKit graphical dialog)
//
// Design principle (see project spec): the application itself NEVER runs as
// administrator. We elevate only the specific, previewed command the student
// has agreed to — never the whole app. Every elevated request carries a
// human-readable Reason shown to the student BEFORE execution for
// transparency and pedagogical value.
package elevate

import (
	"context"
	"runtime"
	"strings"
)

// Request describes one privileged action.
type Request struct {
	// Command is the executable to run (e.g. "winget", "apt-get").
	Command string
	// Args are its arguments.
	Args []string
	// Reason is the short Spanish explanation of WHY this needs admin rights,
	// shown to the student in the confirmation panel before the OS dialog.
	Reason string
}

// Result captures the outcome of an elevated run.
type Result struct {
	// OK is true when the command exited 0.
	OK bool
	// ExitCode is the process exit code (-1 if it never started).
	ExitCode int
	// Stdout / Stderr are the captured streams (best effort per platform).
	Stdout string
	Stderr string
	// Cancelled is true when the student declined the OS elevation dialog.
	Cancelled bool
}

// Preview returns the exact command line that will be executed with
// elevation, so the UI can show it to the student verbatim before asking
// for approval. This is intentionally the human-readable form, not the
// internal wrapper we may use to capture output.
func Preview(r Request) string {
	parts := append([]string{r.Command}, r.Args...)
	quoted := make([]string, len(parts))
	for i, p := range parts {
		if strings.ContainsAny(p, " \t\"") {
			quoted[i] = `"` + strings.ReplaceAll(p, `"`, `\"`) + `"`
		} else {
			quoted[i] = p
		}
	}
	return strings.Join(quoted, " ")
}

// Run executes the request with native elevation. The OS shows its own
// elevation dialog; if the student declines, Result.Cancelled is true and
// err is nil (a declined elevation is a normal outcome, not a failure).
func Run(ctx context.Context, r Request) (Result, error) {
	return runElevated(ctx, r)
}

// Mechanism returns the human name of the elevation mechanism on this OS,
// for display in the UI ("UAC", "PolicyKit (pkexec)", …).
func Mechanism() string {
	switch runtime.GOOS {
	case "windows":
		return "UAC (Control de cuentas de usuario)"
	case "darwin":
		return "Autenticación de administrador de macOS"
	case "linux":
		return "PolicyKit (pkexec)"
	default:
		return "desconocido"
	}
}
