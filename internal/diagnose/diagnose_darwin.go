//go:build darwin
// +build darwin

package diagnose

import (
	"context"
	"os/exec"
	"runtime"
	"strings"
)

// gatherPlatform on macOS checks the CPU virtualization feature via sysctl.
// On Intel Macs, machdep.cpu.features contains "VMX". We also surface an
// architecture warning: this tool + the amd64 box only work on Intel Macs,
// not Apple Silicon.
func gatherPlatform(ctx context.Context) []Probe {
	probes := []Probe{archProbeDarwin()}
	probes = append(probes, virtProbeDarwin(ctx))
	return probes
}

func archProbeDarwin() Probe {
	if runtime.GOARCH == "arm64" {
		return Probe{Key: "arch", Label: "Arquitectura del Mac", Level: Error,
			Value: "Apple Silicon (arm64)",
			Advice: "Este laboratorio usa VirtualBox con una caja amd64, que no es compatible con Apple Silicon. Necesitas un Mac con procesador Intel, o usa la versión portable / otra alternativa."}
	}
	return Probe{Key: "arch", Label: "Arquitectura del Mac", Level: OK, Value: "Intel (x86-64)"}
}

func virtProbeDarwin(ctx context.Context) Probe {
	out, err := exec.CommandContext(ctx, "sysctl", "-n", "machdep.cpu.features").Output()
	if err != nil {
		// On Apple Silicon this sysctl key doesn't exist; arch probe already
		// flagged it. Otherwise report unknown.
		return Probe{Key: "cpu_virt", Label: "Virtualización por hardware (VT-x)", Level: Warn,
			Value: "desconocida", Advice: "No pude leer machdep.cpu.features: " + err.Error()}
	}
	if strings.Contains(strings.ToUpper(string(out)), "VMX") {
		return Probe{Key: "cpu_virt", Label: "Virtualización por hardware (VT-x)", Level: OK,
			Value: "Soportada (VMX presente)"}
	}
	return Probe{Key: "cpu_virt", Label: "Virtualización por hardware (VT-x)", Level: Error,
		Value: "VMX no detectado",
		Advice: "Tu Mac no reporta soporte VMX. En Macs Intel suele venir habilitado de fábrica; si estás dentro de una VM, habilita la virtualización anidada."}
}
