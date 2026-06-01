//go:build windows
// +build windows

package vagrant

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// OpenInteractiveSSH opens a NEW console window with an interactive
// `vagrant ssh` session into the VM (user: vagrant). We write a small .bat
// that cd's to the Vagrant working dir and runs `vagrant ssh`, then launch it
// detached via `cmd /c start`. The window survives the SSH session (pause at
// the end) so the student can read any final message.
func (c *Client) OpenInteractiveSSH() error {
	bat := filepath.Join(os.TempDir(), "bdp-vagrant-ssh.bat")
	content := "@echo off\r\n" +
		"title BDP - Consola SSH a la VM (usuario: vagrant)\r\n" +
		"cd /d \"" + c.WorkDir + "\"\r\n" +
		"echo ================================================================\r\n" +
		"echo   Consola SSH a la VM del laboratorio (usuario: vagrant)\r\n" +
		"echo ================================================================\r\n" +
		"echo.\r\n" +
		"echo  Scripts utiles dentro de la VM:\r\n" +
		"echo    sudo /usr/local/bin/quasar-start.sh   # iniciar el stack\r\n" +
		"echo    sudo /usr/local/bin/quasar-stop.sh    # detener el stack\r\n" +
		"echo    hdfs dfs -ls /                        # explorar HDFS\r\n" +
		"echo    jps                                   # ver procesos Java\r\n" +
		"echo.\r\n" +
		"echo  Escribe 'exit' para cerrar la sesion.\r\n" +
		"echo ----------------------------------------------------------------\r\n" +
		"\"" + c.Bin + "\" ssh\r\n" +
		"echo.\r\n" +
		"echo Sesion SSH terminada. Pulsa una tecla para cerrar.\r\n" +
		"pause >nul\r\n"
	if err := os.WriteFile(bat, []byte(content), 0o644); err != nil {
		return fmt.Errorf("no pude escribir el lanzador SSH: %w", err)
	}
	// start "<title>" cmd /c "<bat>"  → opens a new visible console window.
	cmd := exec.Command("cmd.exe", "/c", "start", "BDP SSH", "cmd", "/c", bat)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("no pude abrir la consola SSH: %w", err)
	}
	return nil
}
