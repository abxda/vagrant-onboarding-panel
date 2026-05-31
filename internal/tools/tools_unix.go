//go:build darwin || linux
// +build darwin linux

package tools

import (
	"os"
	"os/exec"
	"runtime"
)

func vboxManageNames() []string { return []string{"VBoxManage", "vboxmanage"} }
func vagrantNames() []string    { return []string{"vagrant"} }

func packageManagerCandidates() []string {
	if runtime.GOOS == "darwin" {
		return []string{"brew"}
	}
	return []string{"apt-get", "dnf", "pacman"}
}

func vboxKnownPaths() []string {
	if runtime.GOOS == "darwin" {
		return []string{"/usr/local/bin/VBoxManage", "/opt/homebrew/bin/VBoxManage", "/Applications/VirtualBox.app/Contents/MacOS/VBoxManage"}
	}
	return []string{"/usr/bin/VBoxManage", "/usr/local/bin/VBoxManage"}
}

func vagrantKnownPaths() []string {
	if runtime.GOOS == "darwin" {
		return []string{"/usr/local/bin/vagrant", "/opt/homebrew/bin/vagrant", "/opt/vagrant/bin/vagrant"}
	}
	return []string{"/usr/bin/vagrant", "/usr/local/bin/vagrant", "/opt/vagrant/bin/vagrant"}
}

func fileExists(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && !fi.IsDir()
}

// hideWindow is a no-op on Unix (no stray console window).
func hideWindow(cmd *exec.Cmd) {}
