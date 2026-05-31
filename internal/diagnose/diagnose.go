// Package diagnose gathers the read-only system checks the wizard's first
// step needs before any virtualization software is installed: hardware
// virtualization (VT-x / AMD-V), available RAM and disk, and — on Windows —
// hypervisor conflicts (Hyper-V / VBS / Memory Integrity) that degrade or
// disable VirtualBox. Nothing here mutates the system.
package diagnose

import (
	"context"
	"fmt"
	"os"

	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

// Level is a probe's traffic-light severity.
type Level string

const (
	OK    Level = "ok"
	Warn  Level = "warn"
	Error Level = "error"
)

// Probe is one diagnostic result shown as a row in the UI.
type Probe struct {
	Key    string `json:"key"`
	Label  string `json:"label"`
	Level  Level  `json:"level"`
	Value  string `json:"value"`  // the measured value ("16 GB", "Habilitado"…)
	Advice string `json:"advice"` // actionable Spanish guidance when not OK
}

// Report is the full diagnostic outcome.
type Report struct {
	Probes  []Probe `json:"probes"`
	Overall Level   `json:"overall"`
	// CanProceed is false only when a hard blocker (Error) is present.
	CanProceed bool `json:"canProceed"`
}

// Minimum resources the box needs comfortably. The VM is provisioned with
// 4 GB RAM; the box image is ~4.4 GB and the VM disk grows past that.
const (
	minRAMTotalGB  = 8.0  // recommended host RAM; warn below
	minRAMFreeGB   = 4.0  // VM wants 4 GB; warn if less free
	minDiskFreeGB  = 20.0 // box + VM disk headroom; warn below
	hardDiskFreeGB = 10.0 // below this it almost certainly won't fit; error
)

// Run executes all probes and computes the overall verdict.
func Run(ctx context.Context) Report {
	var probes []Probe

	// Platform-specific virtualization + hypervisor checks.
	probes = append(probes, gatherPlatform(ctx)...)

	// Shared RAM / disk checks via gopsutil.
	probes = append(probes, ramProbe())
	probes = append(probes, diskProbe())

	overall := OK
	canProceed := true
	for _, p := range probes {
		if p.Level == Error {
			overall = Error
			canProceed = false
		} else if p.Level == Warn && overall != Error {
			overall = Warn
		}
	}
	return Report{Probes: probes, Overall: overall, CanProceed: canProceed}
}

func ramProbe() Probe {
	vm, err := mem.VirtualMemory()
	if err != nil {
		return Probe{Key: "ram", Label: "Memoria RAM", Level: Warn, Value: "desconocida",
			Advice: "No pude leer la memoria del sistema: " + err.Error()}
	}
	totalGB := float64(vm.Total) / (1 << 30)
	availGB := float64(vm.Available) / (1 << 30)
	val := fmt.Sprintf("%.1f GB totales · %.1f GB disponibles", totalGB, availGB)

	if totalGB < minRAMFreeGB+1 {
		return Probe{Key: "ram", Label: "Memoria RAM", Level: Error, Value: val,
			Advice: fmt.Sprintf("La VM necesita %.0f GB. Tu equipo tiene muy poca RAM total; cierra todo o usa otro equipo.", minRAMFreeGB)}
	}
	if totalGB < minRAMTotalGB || availGB < minRAMFreeGB {
		return Probe{Key: "ram", Label: "Memoria RAM", Level: Warn, Value: val,
			Advice: fmt.Sprintf("La VM reserva %.0f GB. Con menos de %.0f GB disponibles tu equipo puede ir lento mientras corre. Cierra aplicaciones pesadas antes de levantar la VM.", minRAMFreeGB, minRAMFreeGB)}
	}
	return Probe{Key: "ram", Label: "Memoria RAM", Level: OK, Value: val}
}

func diskProbe() Probe {
	path := vmDiskPath()
	usage, err := disk.Usage(path)
	if err != nil {
		return Probe{Key: "disk", Label: "Espacio en disco", Level: Warn, Value: "desconocido",
			Advice: "No pude leer el espacio libre en " + path + ": " + err.Error()}
	}
	freeGB := float64(usage.Free) / (1 << 30)
	val := fmt.Sprintf("%.1f GB libres en %s", freeGB, path)

	if freeGB < hardDiskFreeGB {
		return Probe{Key: "disk", Label: "Espacio en disco", Level: Error, Value: val,
			Advice: fmt.Sprintf("La caja ocupa ~4.4 GB y la VM crece más. Necesitas al menos %.0f GB libres. Libera espacio antes de continuar.", minDiskFreeGB)}
	}
	if freeGB < minDiskFreeGB {
		return Probe{Key: "disk", Label: "Espacio en disco", Level: Warn, Value: val,
			Advice: fmt.Sprintf("Se recomienda %.0f GB libres (caja ~4.4 GB + crecimiento de la VM). Vas justo; considera liberar espacio.", minDiskFreeGB)}
	}
	return Probe{Key: "disk", Label: "Espacio en disco", Level: OK, Value: val}
}

// vmDiskPath returns the path whose free space matters — where Vagrant/
// VirtualBox will store the VM. We use the user's home directory drive,
// falling back to the current directory.
func vmDiskPath() string {
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		return home
	}
	if cwd, err := os.Getwd(); err == nil {
		return cwd
	}
	return "."
}
