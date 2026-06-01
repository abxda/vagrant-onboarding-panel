//go:build darwin || linux
// +build darwin linux

package vagrant

import "os/exec"

// hideConsole es no-op en Unix (no hay consola que ocultar).
func hideConsole(cmd *exec.Cmd) {}
