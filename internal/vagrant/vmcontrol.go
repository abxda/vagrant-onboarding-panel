package vagrant

import (
	"context"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// VMName es el nombre FIJO que el Vagrantfile asigna a la máquina virtual
// (vb.name). Es la clave para controlar la VM por nombre con VBoxManage SIN
// necesitar el directorio del Vagrantfile — VirtualBox lleva un registro
// global de máquinas. Así, cualquier proceso (este panel, el portable, el
// meta-launcher) puede apagar la VM aunque no sepa desde qué carpeta-semilla
// se levantó.
const VMName = "BDP-BigDataLab"

// vboxManage localiza el ejecutable VBoxManage. PATH primero; si no, rutas
// conocidas por SO (tras instalar, VBoxManage no siempre está en el PATH del
// proceso actual).
func vboxManage() string {
	if p, err := exec.LookPath(vboxManageBin()); err == nil {
		return p
	}
	for _, c := range vboxKnownPaths() {
		if fileExists(c) {
			return c
		}
	}
	return vboxManageBin() // último recurso; fallará con mensaje claro
}

func vboxManageBin() string {
	if runtime.GOOS == "windows" {
		return "VBoxManage.exe"
	}
	return "VBoxManage"
}

func vboxKnownPaths() []string {
	if runtime.GOOS == "windows" {
		return []string{
			`C:\Program Files\Oracle\VirtualBox\VBoxManage.exe`,
			`C:\Program Files\Oracle\VirtualBox\VBoxManage`,
		}
	}
	if runtime.GOOS == "darwin" {
		return []string{"/usr/local/bin/VBoxManage", "/opt/homebrew/bin/VBoxManage", "/Applications/VirtualBox.app/Contents/MacOS/VBoxManage"}
	}
	return []string{"/usr/bin/VBoxManage", "/usr/local/bin/VBoxManage"}
}

// VMIsRunning reporta si la VM del laboratorio está en ejecución, consultando
// VirtualBox por NOMBRE (no por directorio). Funciona desde cualquier carpeta.
func VMIsRunning(ctx context.Context) bool {
	out, err := vboxOutput(ctx, "list", "runningvms")
	if err != nil {
		return false
	}
	return strings.Contains(out, `"`+VMName+`"`)
}

// VMExists reporta si la VM existe registrada en VirtualBox (corriendo o no).
func VMExists(ctx context.Context) bool {
	out, err := vboxOutput(ctx, "list", "vms")
	if err != nil {
		return false
	}
	return strings.Contains(out, `"`+VMName+`"`)
}

// PowerButton manda la señal ACPI de apagado (equivale al botón físico de
// power): el SO invitado (Debian) cierra ordenadamente. Apagado LIMPIO por
// NOMBRE, sin necesitar el directorio del Vagrantfile. No espera a que termine.
func PowerButton(ctx context.Context) error {
	_, err := vboxOutput(ctx, "controlvm", VMName, "acpipowerbutton")
	return err
}

// WaitPoweredOff espera (hasta timeout) a que la VM deje de estar corriendo,
// tras un PowerButton. Devuelve true si se apagó a tiempo.
func WaitPoweredOff(ctx context.Context, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if !VMIsRunning(ctx) {
			return true
		}
		time.Sleep(2 * time.Second)
	}
	return !VMIsRunning(ctx)
}

func vboxOutput(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, vboxManage(), args...)
	hideConsole(cmd) // platform hook: oculta la consola en Windows
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func fileExists(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && !fi.IsDir()
}
