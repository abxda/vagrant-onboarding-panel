//go:build darwin || linux
// +build darwin linux

package vagrant

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// OpenInteractiveSSH opens a new terminal window with an interactive
// `vagrant ssh` session. On macOS it writes a .command script and opens it with
// Terminal; on Linux it tries the common terminal emulators.
func (c *Client) OpenInteractiveSSH() error {
	script := filepath.Join(os.TempDir(), "bdp-vagrant-ssh.command")
	body := "#!/bin/bash\n" +
		"cd \"" + c.WorkDir + "\"\n" +
		"echo '================================================================'\n" +
		"echo '  Consola SSH a la VM del laboratorio (usuario: vagrant)'\n" +
		"echo '================================================================'\n" +
		"echo ''\n" +
		"echo ' Scripts utiles dentro de la VM:'\n" +
		"echo '   sudo /usr/local/bin/quasar-start.sh   # iniciar el stack'\n" +
		"echo '   sudo /usr/local/bin/quasar-stop.sh    # detener el stack'\n" +
		"echo '   hdfs dfs -ls /                        # explorar HDFS'\n" +
		"echo ''\n" +
		"\"" + c.Bin + "\" ssh\n" +
		"echo ''\n" +
		"echo 'Sesion SSH terminada.'\n"
	if err := os.WriteFile(script, []byte(body), 0o755); err != nil {
		return fmt.Errorf("no pude escribir el lanzador SSH: %w", err)
	}

	if runtime.GOOS == "darwin" {
		if err := exec.Command("open", "-a", "Terminal", script).Start(); err != nil {
			return fmt.Errorf("no pude abrir Terminal: %w", err)
		}
		return nil
	}

	// Linux: try common terminal emulators.
	for _, term := range []string{"x-terminal-emulator", "gnome-terminal", "konsole", "xterm"} {
		if _, err := exec.LookPath(term); err == nil {
			if err := exec.Command(term, "-e", "bash", script).Start(); err == nil {
				return nil
			}
		}
	}
	return fmt.Errorf("no encontré un emulador de terminal. Abre uno y ejecuta: cd %s && %s ssh", c.WorkDir, c.Bin)
}
