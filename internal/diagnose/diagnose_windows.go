//go:build windows
// +build windows

package diagnose

import (
	"context"

	"github.com/yusufpapurcu/wmi"
	"golang.org/x/sys/windows/registry"
)

// win32Processor mirrors the WMI fields we read. Pointers so we can tell
// "false" apart from "not returned".
type win32Processor struct {
	Name                          string
	VirtualizationFirmwareEnabled *bool
}

type win32ComputerSystem struct {
	HypervisorPresent *bool
}

// gatherPlatform runs the Windows-specific virtualization + hypervisor probes.
//
// Decision tree for VT-x:
//   - If a hypervisor is already present (Hyper-V / VBS / WSL2), Windows owns
//     VT-x and VirtualBox will be slow or fail. We report that as the conflict
//     and cannot meaningfully read the firmware flag (Windows hides it).
//   - Otherwise we read Win32_Processor.VirtualizationFirmwareEnabled: true →
//     OK, false → VT-x disabled in BIOS (hard blocker with BIOS advice).
func gatherPlatform(ctx context.Context) []Probe {
	hyperPresent := hypervisorPresent()

	var probes []Probe
	probes = append(probes, virtProbe(hyperPresent))
	probes = append(probes, hypervisorProbe(hyperPresent))
	probes = append(probes, memoryIntegrityProbe())
	return probes
}

func hypervisorPresent() bool {
	var cs []win32ComputerSystem
	if err := wmi.Query("SELECT HypervisorPresent FROM Win32_ComputerSystem", &cs); err != nil {
		return false
	}
	if len(cs) > 0 && cs[0].HypervisorPresent != nil {
		return *cs[0].HypervisorPresent
	}
	return false
}

func virtProbe(hyperPresent bool) Probe {
	if hyperPresent {
		// A hypervisor is running; VT-x is owned by it. The hardware almost
		// certainly supports virtualization (the hypervisor needs it), so we
		// don't flag the CPU itself — the conflict probe covers the real issue.
		return Probe{
			Key:   "cpu_virt",
			Label: "Virtualización por hardware (VT-x/AMD-V)",
			Level: OK,
			Value: "Soportada (en uso por un hipervisor de Windows)",
		}
	}

	var procs []win32Processor
	if err := wmi.Query("SELECT Name, VirtualizationFirmwareEnabled FROM Win32_Processor", &procs); err != nil {
		return Probe{Key: "cpu_virt", Label: "Virtualización por hardware (VT-x/AMD-V)", Level: Warn,
			Value: "desconocida", Advice: "No pude consultar el procesador vía WMI: " + err.Error()}
	}
	if len(procs) == 0 || procs[0].VirtualizationFirmwareEnabled == nil {
		return Probe{Key: "cpu_virt", Label: "Virtualización por hardware (VT-x/AMD-V)", Level: Warn,
			Value: "desconocida",
			Advice: "Windows no reportó el estado de virtualización. Verifica en el Administrador de tareas → Rendimiento → CPU que 'Virtualización' diga 'Habilitada'."}
	}
	if *procs[0].VirtualizationFirmwareEnabled {
		return Probe{Key: "cpu_virt", Label: "Virtualización por hardware (VT-x/AMD-V)", Level: OK,
			Value: "Habilitada en firmware"}
	}
	return Probe{Key: "cpu_virt", Label: "Virtualización por hardware (VT-x/AMD-V)", Level: Error,
		Value: "Deshabilitada en BIOS/UEFI",
		Advice: "Entra a la BIOS/UEFI de tu equipo y activa 'Intel VT-x' (o 'AMD-V' / 'SVM Mode'). Suele estar en Advanced → CPU Configuration. Sin esto VirtualBox no puede ejecutar máquinas de 64 bits."}
}

func hypervisorProbe(hyperPresent bool) Probe {
	if !hyperPresent {
		return Probe{Key: "hypervisor", Label: "Conflicto de hipervisor (Hyper-V / VBS)", Level: OK,
			Value: "Sin hipervisor activo"}
	}
	return Probe{Key: "hypervisor", Label: "Conflicto de hipervisor (Hyper-V / VBS)", Level: Warn,
		Value: "Hay un hipervisor activo (Hyper-V, WSL2 o Seguridad basada en virtualización)",
		Advice: "Un hipervisor de Windows tiene tomado VT-x, lo que hace que VirtualBox vaya muy lento (ícono de tortuga) o falle. Para liberarlo: 1) ejecuta como administrador  bcdedit /set hypervisorlaunchtype off  2) desactiva 'Integridad de memoria' en Seguridad de Windows  3) en 'Activar o desactivar características de Windows' desmarca Hyper-V y 'Plataforma de hipervisor de Windows'  4) reinicia en frío (apagar y encender)."}
}

// memoryIntegrityProbe reads the Core Isolation > Memory Integrity setting,
// which silently pulls in VBS/the hypervisor and is a very common cause of
// VirtualBox slowdowns on Windows 11.
func memoryIntegrityProbe() Probe {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`SYSTEM\CurrentControlSet\Control\DeviceGuard\Scenarios\HypervisorEnforcedCodeIntegrity`,
		registry.QUERY_VALUE)
	if err != nil {
		// Key absent usually means the feature was never enabled → fine.
		return Probe{Key: "memory_integrity", Label: "Integridad de memoria (Core Isolation)", Level: OK,
			Value: "Desactivada o no configurada"}
	}
	defer k.Close()
	enabled, _, err := k.GetIntegerValue("Enabled")
	if err != nil {
		return Probe{Key: "memory_integrity", Label: "Integridad de memoria (Core Isolation)", Level: OK,
			Value: "No configurada"}
	}
	if enabled == 1 {
		return Probe{Key: "memory_integrity", Label: "Integridad de memoria (Core Isolation)", Level: Warn,
			Value: "Activada",
			Advice: "La 'Integridad de memoria' activa la virtualización de Windows (VBS), que compite con VirtualBox. Desactívala en Seguridad de Windows → Seguridad del dispositivo → Aislamiento del núcleo → Integridad de memoria, y reinicia."}
	}
	return Probe{Key: "memory_integrity", Label: "Integridad de memoria (Core Isolation)", Level: OK,
		Value: "Desactivada"}
}
