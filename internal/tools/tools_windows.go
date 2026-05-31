//go:build windows
// +build windows

package tools

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func vboxManageNames() []string { return []string{"VBoxManage.exe", "VBoxManage"} }
func vagrantNames() []string    { return []string{"vagrant.exe", "vagrant"} }

func packageManagerCandidates() []string { return []string{"winget", "choco"} }

// vboxKnownPaths returns the standard VirtualBox install locations on Windows.
// The installer sets VBOX_MSI_INSTALL_PATH / VBOX_INSTALL_PATH in the machine
// environment, but those aren't visible to an already-running process, so we
// also hard-check Program Files.
func vboxKnownPaths() []string {
	var out []string
	for _, env := range []string{"VBOX_MSI_INSTALL_PATH", "VBOX_INSTALL_PATH"} {
		if v := os.Getenv(env); v != "" {
			out = append(out, filepath.Join(v, "VBoxManage.exe"))
		}
	}
	for _, pf := range []string{os.Getenv("ProgramFiles"), os.Getenv("ProgramW6432"), `C:\Program Files`} {
		if pf != "" {
			out = append(out, filepath.Join(pf, "Oracle", "VirtualBox", "VBoxManage.exe"))
		}
	}
	return out
}

func vagrantKnownPaths() []string {
	var out []string
	for _, base := range []string{
		`C:\HashiCorp\Vagrant\bin`,
		filepath.Join(os.Getenv("ProgramFiles"), "Vagrant", "bin"),
		filepath.Join(os.Getenv("ProgramW6432"), "Vagrant", "bin"),
	} {
		if base != "" {
			out = append(out, filepath.Join(base, "vagrant.exe"))
		}
	}
	return out
}

func fileExists(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && !fi.IsDir()
}

// hideWindow prevents a console window flashing when we probe a tool version.
func hideWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
}
