//go:build darwin || linux
// +build darwin linux

package runner

import "os/exec"

// applySysProcAttr is a no-op on Unix — there is no stray console window to
// hide. Defined so runner.go stays platform-agnostic.
func applySysProcAttr(cmd *exec.Cmd) {}
