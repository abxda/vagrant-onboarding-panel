// Package tools detects whether VirtualBox and Vagrant are installed and at
// what version, and which native package manager is available to install
// them. Detection searches PATH plus OS-specific well-known install
// locations — important on Windows, where right after an install the new
// binary is NOT yet on the current process's PATH, but lives at a
// predictable path (C:\Program Files\Oracle\VirtualBox, etc.).
package tools

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"time"
)

// Info describes a detected (or missing) tool.
type Info struct {
	Installed bool   `json:"installed"`
	Version   string `json:"version"`
	Path      string `json:"path"`
}

// DetectVirtualBox locates VBoxManage and reads its version.
func DetectVirtualBox(ctx context.Context) Info {
	return detect(ctx, vboxManageNames(), vboxKnownPaths(), []string{"--version"}, parseVBoxVersion)
}

// DetectVagrant locates the vagrant binary and reads its version.
func DetectVagrant(ctx context.Context) Info {
	return detect(ctx, vagrantNames(), vagrantKnownPaths(), []string{"--version"}, parseVagrantVersion)
}

// PackageManager returns the native package manager available on this host
// for installing software ("winget", "brew", "apt-get"), and whether one was
// found.
func PackageManager() (string, bool) {
	for _, pm := range packageManagerCandidates() {
		if _, err := exec.LookPath(pm); err == nil {
			return pm, true
		}
	}
	return "", false
}

// detect tries each binary name on PATH, then each known absolute path,
// running the version args and parsing the output with parse.
func detect(ctx context.Context, names, knownPaths, versionArgs []string, parse func(string) string) Info {
	var candidates []string
	for _, n := range names {
		if p, err := exec.LookPath(n); err == nil {
			candidates = append(candidates, p)
		}
	}
	candidates = append(candidates, knownPaths...)

	for _, bin := range candidates {
		if !fileExists(bin) {
			continue
		}
		out, ok := runVersion(ctx, bin, versionArgs)
		if !ok {
			continue
		}
		return Info{Installed: true, Version: parse(out), Path: bin}
	}
	return Info{Installed: false}
}

func runVersion(ctx context.Context, bin string, args []string) (string, bool) {
	cctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(cctx, bin, args...)
	hideWindow(cmd)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		// Some tools print version then exit non-zero; still use the output if
		// it looks like a version string.
		s := out.String()
		if strings.TrimSpace(s) == "" {
			return "", false
		}
		return s, true
	}
	return out.String(), true
}

func parseVBoxVersion(s string) string {
	// "7.0.18r162988\n" → "7.0.18"
	s = strings.TrimSpace(s)
	if i := strings.IndexByte(s, 'r'); i > 0 {
		return s[:i]
	}
	return firstLine(s)
}

func parseVagrantVersion(s string) string {
	// "Vagrant 2.4.1\n" → "2.4.1"
	line := firstLine(s)
	fields := strings.Fields(line)
	if len(fields) >= 2 && strings.EqualFold(fields[0], "Vagrant") {
		return fields[1]
	}
	return line
}

func firstLine(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexAny(s, "\r\n"); i >= 0 {
		return strings.TrimSpace(s[:i])
	}
	return s
}
