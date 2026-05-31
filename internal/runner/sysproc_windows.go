//go:build windows
// +build windows

package runner

import (
	"os/exec"
	"syscall"
)

// applySysProcAttr hides the console window that would otherwise flash when
// the GUI app spawns a console subprocess (vboxmanage, vagrant, …).
func applySysProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000} // CREATE_NO_WINDOW
}
