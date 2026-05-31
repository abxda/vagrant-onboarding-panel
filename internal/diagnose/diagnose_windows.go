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
		return Probe{Key: "hypervisor", Label: "Hipervisor de Windows (Hyper-V / VBS)", Level: OK,
			Value: "Sin hipervisor activo · VirtualBox correrá a máxima velocidad"}
	}
	// Strategy chosen for students: COEXIST with the Windows hypervisor. We do
	// NOT disable anything. VirtualBox 7 detects an active hypervisor and runs
	// on top of it (Hyper-V backend), shown by a green turtle icon. The VM
	// works — just slower. This message reassures rather than alarms, and never
	// touches the student's security configuration.
	return Probe{Key: "hypervisor", Label: "Hipervisor de Windows (Hyper-V / VBS)", Level: Warn,
		Value: "Activo · VirtualBox funcionará en modo compatibilidad (más lento)",
		Advice: "Tu Windows tiene activada la seguridad por virtualización (VBS / Integridad de memoria). Es NORMAL en Windows 11 y NO lo vamos a desactivar: respetamos tu equipo y tu seguridad. VirtualBox 7 corre por encima del hipervisor de Windows; verás un ícono de tortuga verde en la VM — es esperado, no es un error. La máquina virtual funciona igual, solo un poco más lenta. (Opcional y avanzado: si más adelante quisieras máxima velocidad, existe la vía de desactivar el hipervisor, pero requiere reiniciar y reduce temporalmente la seguridad; por eso NO es el camino por defecto.)"}
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
		return Probe{Key: "memory_integrity", Label: "Integridad de memoria (Core Isolation)", Level: OK,
			Value: "Activada · la dejamos como está",
			Advice: "Tienes la 'Integridad de memoria' activada (es una buena protección). No la tocamos: VirtualBox funcionará sobre ella en modo compatibilidad. Solo te informamos para que sepas por qué la VM puede ir un poco más lenta."}
	}
	return Probe{Key: "memory_integrity", Label: "Integridad de memoria (Core Isolation)", Level: OK,
		Value: "Desactivada"}
}
