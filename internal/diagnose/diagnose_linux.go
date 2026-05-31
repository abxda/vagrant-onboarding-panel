//go:build linux
// +build linux

package diagnose

import (
	"context"
	"os"
	"strings"
)

// gatherPlatform on Linux reads /proc/cpuinfo for the vmx (Intel) or svm
// (AMD) CPU flags. There is no Hyper-V conflict to worry about; KVM can
// coexist with VirtualBox in most cases, so we only note it.
func gatherPlatform(ctx context.Context) []Probe {
	return []Probe{virtProbeLinux(), kvmNoteLinux()}
}

func virtProbeLinux() Probe {
	data, err := os.ReadFile("/proc/cpuinfo")
	if err != nil {
		return Probe{Key: "cpu_virt", Label: "Virtualización por hardware (VT-x/AMD-V)", Level: Warn,
			Value: "desconocida", Advice: "No pude leer /proc/cpuinfo: " + err.Error()}
	}
	text := string(data)
	if strings.Contains(text, "vmx") {
		return Probe{Key: "cpu_virt", Label: "Virtualización por hardware (VT-x)", Level: OK,
			Value: "Soportada (flag vmx presente)"}
	}
	if strings.Contains(text, "svm") {
		return Probe{Key: "cpu_virt", Label: "Virtualización por hardware (AMD-V)", Level: OK,
			Value: "Soportada (flag svm presente)"}
	}
	return Probe{Key: "cpu_virt", Label: "Virtualización por hardware (VT-x/AMD-V)", Level: Error,
		Value: "No detectada",
		Advice: "No encontré los flags 'vmx' ni 'svm' en /proc/cpuinfo. Activa la virtualización (VT-x/AMD-V) en la BIOS/UEFI. Si estás dentro de otra VM, habilita la virtualización anidada."}
}

func kvmNoteLinux() Probe {
	if _, err := os.Stat("/dev/kvm"); err == nil {
		return Probe{Key: "hypervisor", Label: "Módulo KVM", Level: OK,
			Value: "/dev/kvm presente (normalmente convive con VirtualBox)"}
	}
	return Probe{Key: "hypervisor", Label: "Módulo KVM", Level: OK,
		Value: "KVM no cargado"}
}
