// Package vagrant wraps the Vagrant CLI for the wizard's box / up / services
// steps. All commands run UNELEVATED (Vagrant + VirtualBox run as the user)
// and stream their output to the shared log sink via the runner.
package vagrant

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/abxda/vagrant-onboarding-panel/internal/runner"
)

// BoxName is the lab box published on the HCP Vagrant Registry.
const BoxName = "abxda/big-data-lab"

// Provider is fixed to VirtualBox for this project.
const Provider = "virtualbox"

// Client drives Vagrant in a specific working directory (where the generated
// Vagrantfile lives).
type Client struct {
	Bin     string // absolute path to the vagrant executable
	WorkDir string // directory containing the Vagrantfile
	r       *runner.Runner
}

func NewClient(bin, workDir string, r *runner.Runner) *Client {
	return &Client{Bin: bin, WorkDir: workDir, r: r}
}

// BoxAdded reports whether the lab box is already registered locally.
func (c *Client) BoxAdded(ctx context.Context) (bool, error) {
	res, err := c.r.RunDir(ctx, 30*time.Second, c.WorkDir, c.Bin, "box", "list")
	if err != nil {
		return false, err
	}
	return strings.Contains(res.Stdout, BoxName), nil
}

// AddBox downloads and registers the lab box (~4.4 GB). Long timeout.
func (c *Client) AddBox(ctx context.Context) (runner.Result, error) {
	return c.r.RunDir(ctx, 90*time.Minute, c.WorkDir, c.Bin,
		"box", "add", BoxName, "--provider", Provider)
}

// EnsureVagrantfile writes the Vagrantfile (and the exercise dir) into WorkDir
// if not already present. exerciseSrcWritten lists files materialised under
// WorkDir/ejercicio_01 (for the file provisioner).
func (c *Client) EnsureVagrantfile() error {
	vf := filepath.Join(c.WorkDir, "Vagrantfile")
	content := vagrantfile()
	return os.WriteFile(vf, []byte(content), 0o644)
}

// Up boots the VM (downloads the box first if needed). Long timeout because the
// first boot provisions and the host may be running VirtualBox over Hyper-V.
func (c *Client) Up(ctx context.Context) (runner.Result, error) {
	return c.r.RunDir(ctx, 60*time.Minute, c.WorkDir, c.Bin, "up", "--provider", Provider)
}

// State returns the VM state ("running", "poweroff", "not_created", …) parsed
// from `vagrant status --machine-readable`.
func (c *Client) State(ctx context.Context) (string, error) {
	res, err := c.r.RunDir(ctx, 60*time.Second, c.WorkDir, c.Bin, "status", "--machine-readable")
	if err != nil {
		return "", err
	}
	// machine-readable lines: timestamp,target,type,data...
	for _, line := range strings.Split(res.Stdout, "\n") {
		parts := strings.Split(strings.TrimSpace(line), ",")
		if len(parts) >= 4 && parts[2] == "state" {
			return parts[3], nil
		}
	}
	return "unknown", nil
}

// Upload copies a host path into the guest via `vagrant upload` (uses SCP, so
// it does NOT require VirtualBox Guest Additions / synced folders).
func (c *Client) Upload(ctx context.Context, hostPath, guestPath string) (runner.Result, error) {
	return c.r.RunDir(ctx, 5*time.Minute, c.WorkDir, c.Bin, "upload", hostPath, guestPath)
}

// SSH runs a single shell command inside the VM via `vagrant ssh -c`.
func (c *Client) SSH(ctx context.Context, timeout time.Duration, command string) (runner.Result, error) {
	return c.r.RunDir(ctx, timeout, c.WorkDir, c.Bin, "ssh", "-c", command)
}

// Halt powers off the VM gracefully.
func (c *Client) Halt(ctx context.Context) (runner.Result, error) {
	return c.r.RunDir(ctx, 5*time.Minute, c.WorkDir, c.Bin, "halt")
}

// DefaultWorkDir returns the lab working directory under the user's home.
func DefaultWorkDir() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		if cwd, e := os.Getwd(); e == nil {
			home = cwd
		} else {
			home = "."
		}
	}
	return filepath.Join(home, "bdp-vagrant-lab")
}

// vagrantfile returns the generated Vagrantfile. Ports mirror the box's
// services (Jupyter 8888, HDFS UI 9870, Elasticsearch 9200). The exercise is
// uploaded separately via `vagrant upload` (more robust than synced folders,
// which need Guest Additions).
func vagrantfile() string {
	return fmt.Sprintf(`# -*- mode: ruby -*-
# vi: set ft=ruby :
# Generado por el Panel de Onboarding (Plan B) — Dr. Abel Coronado.
# Laboratorio de Big Data sobre Vagrant + VirtualBox.

Vagrant.configure("2") do |config|
  config.vm.box = "%s"

  # Servicios de la caja, accesibles desde el host:
  config.vm.network "forwarded_port", guest: 8888, host: 8888 # Jupyter Lab
  config.vm.network "forwarded_port", guest: 9870, host: 9870 # HDFS NameNode UI
  config.vm.network "forwarded_port", guest: 9200, host: 9200 # Elasticsearch API

  config.vm.provider "%s" do |vb|
    vb.name   = "BDP-BigDataLab"
    vb.memory = 4096
    vb.cpus   = 4
  end

  # Mensaje al arrancar (informativo para el alumno).
  config.vm.post_up_message = "Laboratorio de Big Data listo. Servicios en :8888 (Jupyter), :9870 (HDFS), :9200 (Elasticsearch)."
end
`, BoxName, Provider)
}
