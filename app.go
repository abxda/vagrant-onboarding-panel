package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/abxda/vagrant-onboarding-panel/internal/diagnose"
	"github.com/abxda/vagrant-onboarding-panel/internal/elevate"
	"github.com/abxda/vagrant-onboarding-panel/internal/logsink"
	"github.com/abxda/vagrant-onboarding-panel/internal/state"
	"github.com/abxda/vagrant-onboarding-panel/internal/wizard"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the Wails-bound application object.
type App struct {
	ctx   context.Context
	sink  *logsink.Sink
	state *state.State
}

// NewApp constructs the App with its log sink.
func NewApp() *App {
	return &App{
		sink: logsink.New(4000),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.state = state.Load(stateFilePath())
	// Forward every log line to the frontend as a Wails event.
	a.sink.OnNew(func(ln logsink.Line) {
		wruntime.EventsEmit(a.ctx, "log", ln)
	})
	a.sink.Emit("INFO", "Panel de onboarding Vagrant iniciado.")
	a.sink.Emit("INFO", fmt.Sprintf("Plataforma: %s/%s · Elevación: %s", runtime.GOOS, runtime.GOARCH, elevate.Mechanism()))
}

func (a *App) domReady(ctx context.Context)         {}
func (a *App) beforeClose(ctx context.Context) bool { return false }
func (a *App) shutdown(ctx context.Context)         {}

// --- types returned to the frontend -------------------------------------

// StepView is a wizard step plus its current persisted status.
type StepView struct {
	wizard.Step
	Status string `json:"status"`
}

// EnvInfo is shown in the footer / about.
type EnvInfo struct {
	OS         string `json:"os"`
	Arch       string `json:"arch"`
	Mechanism  string `json:"mechanism"`
	Author     string `json:"author"`
	AppVersion string `json:"appVersion"`
}

// ActionResult is returned by step actions / elevation tests.
type ActionResult struct {
	OK        bool   `json:"ok"`
	Cancelled bool   `json:"cancelled"`
	Message   string `json:"message"`
	Detail    string `json:"detail"`
}

// --- bound methods ------------------------------------------------------

// GetEnvInfo returns static environment info for the UI.
func (a *App) GetEnvInfo() EnvInfo {
	return EnvInfo{
		OS:         runtime.GOOS,
		Arch:       runtime.GOARCH,
		Mechanism:  elevate.Mechanism(),
		Author:     "Dr. Abel Coronado",
		AppVersion: appVersion,
	}
}

// GetSteps returns all wizard steps with their persisted status.
func (a *App) GetSteps() []StepView {
	steps := wizard.Steps()
	out := make([]StepView, len(steps))
	for i, s := range steps {
		out[i] = StepView{Step: s, Status: a.state.Status(string(s.ID))}
	}
	return out
}

// GetLogSnapshot returns the current log buffer (for initial paint).
func (a *App) GetLogSnapshot() []logsink.Line { return a.sink.Snapshot() }

// ClearLog empties the live log.
func (a *App) ClearLog() { a.sink.Clear() }

// PreviewElevation returns the exact command line a step would run elevated,
// so the UI can show it BEFORE asking for OS approval. Returns "" for steps
// that don't elevate.
func (a *App) PreviewElevation(stepID string) string {
	req, ok := a.elevatedRequestFor(wizard.StepID(stepID))
	if !ok {
		return ""
	}
	return elevate.Preview(req)
}

// TestElevation is the CP1 acceptance check: run a trivial command with
// native OS elevation and confirm we actually got admin/root. Proves the
// elevation plumbing works on this platform before any real install.
func (a *App) TestElevation() ActionResult {
	a.sink.Emit("INFO", strings.Repeat("─", 56))
	a.sink.Emit("INFO", "Prueba de elevación de privilegios ("+elevate.Mechanism()+")")

	req := elevationProbe()
	a.sink.Emit("INFO", "Se ejecutará (elevado): "+elevate.Preview(req))
	a.sink.Emit("INFO", "Aprueba el diálogo del sistema cuando aparezca…")

	ctx, cancel := context.WithTimeout(a.ctx, 90*time.Second)
	defer cancel()
	res, err := elevate.Run(ctx, req)
	if err != nil {
		a.sink.Emit("ERROR", "Error en la elevación: "+err.Error())
		return ActionResult{OK: false, Message: "Falló la elevación", Detail: err.Error()}
	}
	if res.Cancelled {
		a.sink.Emit("WARN", "El usuario canceló el diálogo de elevación.")
		return ActionResult{OK: false, Cancelled: true, Message: "Cancelaste el diálogo de elevación. No se hizo ningún cambio."}
	}
	elevated := probeConfirmsElevation(res.Stdout)
	if elevated {
		a.sink.Emit("INFO", "✓ Elevación confirmada: el comando corrió con privilegios de administrador.")
		return ActionResult{OK: true, Message: "Elevación funcionando correctamente.", Detail: strings.TrimSpace(res.Stdout)}
	}
	a.sink.Emit("WARN", "El comando corrió pero NO se detectaron privilegios elevados.")
	a.sink.Emit("INFO", "Salida: "+strings.TrimSpace(res.Stdout))
	return ActionResult{OK: false, Message: "El comando corrió pero sin privilegios elevados.", Detail: strings.TrimSpace(res.Stdout)}
}

// GetDiagnostics runs the read-only system diagnostic and returns the full
// report (also streamed to the log). Bound for the frontend to render the
// probe table in the Diagnóstico step.
func (a *App) GetDiagnostics() diagnose.Report {
	a.sink.Emit("INFO", strings.Repeat("─", 56))
	a.sink.Emit("INFO", "Diagnóstico del sistema (solo lectura)…")
	ctx, cancel := context.WithTimeout(a.ctx, 30*time.Second)
	defer cancel()
	rep := diagnose.Run(ctx)
	for _, p := range rep.Probes {
		lvl := "INFO"
		if p.Level == diagnose.Warn {
			lvl = "WARN"
		} else if p.Level == diagnose.Error {
			lvl = "ERROR"
		}
		a.sink.Emit(lvl, fmt.Sprintf("%s: %s", p.Label, p.Value))
		if p.Advice != "" && p.Level != diagnose.OK {
			a.sink.Emit(lvl, "  → "+p.Advice)
		}
	}
	// Persist the step status from the overall verdict.
	st := string(wizard.StatusOK)
	switch rep.Overall {
	case diagnose.Warn:
		st = string(wizard.StatusWarn)
	case diagnose.Error:
		st = string(wizard.StatusError)
	}
	a.state.SetStatus(string(wizard.StepDiagnostico), st)
	a.emitStepStatus(string(wizard.StepDiagnostico), st)
	return rep
}

// CheckStep re-checks a step's status. The diagnostico step runs the real
// diagnostic; later steps return their persisted value (real detection for
// those arrives in CP3+).
func (a *App) CheckStep(stepID string) string {
	if stepID == string(wizard.StepDiagnostico) {
		return string(a.GetDiagnostics().Overall) // ok|warn|error align with wizard.Status*
	}
	a.sink.Emit("INFO", "Verificando paso: "+stepID+" (verificación real llega en el siguiente checkpoint)")
	st := a.state.Status(stepID)
	if st == "" {
		st = string(wizard.StatusUnknown)
	}
	return st
}

// RunStep executes a step's action. CP1: mocked for non-elevated steps; the
// elevated steps route through the real elevation path so the OS dialog is
// exercised end-to-end (the wrapped command is a harmless probe for now).
func (a *App) RunStep(stepID string) ActionResult {
	id := wizard.StepID(stepID)

	// Step 1 is the real read-only diagnostic — no mock, no elevation.
	if id == wizard.StepDiagnostico {
		rep := a.GetDiagnostics()
		switch rep.Overall {
		case diagnose.Error:
			return ActionResult{OK: false, Message: "El diagnóstico encontró un bloqueo. Revisa los puntos en rojo antes de continuar."}
		case diagnose.Warn:
			return ActionResult{OK: true, Message: "Diagnóstico con avisos. Puedes continuar, pero revisa los puntos en amarillo."}
		default:
			return ActionResult{OK: true, Message: "Diagnóstico OK. Tu equipo está listo para virtualizar."}
		}
	}

	a.sink.Emit("INFO", strings.Repeat("─", 56))
	a.sink.Emit("INFO", "Ejecutando paso: "+stepID)
	a.state.SetStatus(stepID, string(wizard.StatusRunning))
	a.emitStepStatus(stepID, string(wizard.StatusRunning))

	if req, ok := a.elevatedRequestFor(id); ok {
		a.sink.Emit("INFO", "Este paso requiere administrador. Se ejecutará (elevado):")
		a.sink.Emit("INFO", "  "+elevate.Preview(req))
		a.sink.Emit("WARN", "CP1: se ejecuta una versión de prueba inofensiva; la instalación real llega en el siguiente checkpoint.")
		ctx, cancel := context.WithTimeout(a.ctx, 90*time.Second)
		defer cancel()
		res, err := elevate.Run(ctx, elevationProbe())
		if err != nil || res.Cancelled || !res.OK {
			a.state.SetStatus(stepID, string(wizard.StatusError))
			a.emitStepStatus(stepID, string(wizard.StatusError))
			msg := "No se completó la elevación."
			if res.Cancelled {
				msg = "Cancelaste el diálogo de elevación."
			}
			return ActionResult{OK: false, Cancelled: res.Cancelled, Message: msg}
		}
	} else {
		// Mocked non-elevated work with a little simulated progress.
		for i := 1; i <= 3; i++ {
			a.sink.Emit("INFO", fmt.Sprintf("  … trabajo simulado %d/3", i))
			time.Sleep(250 * time.Millisecond)
		}
	}

	a.state.SetStatus(stepID, string(wizard.StatusOK))
	a.emitStepStatus(stepID, string(wizard.StatusOK))
	a.sink.Emit("INFO", "✓ Paso completado (CP1 simulado): "+stepID)
	return ActionResult{OK: true, Message: "Paso completado."}
}

// ResetWizard clears all persisted progress.
func (a *App) ResetWizard() {
	a.state.Reset()
	a.sink.Emit("INFO", "Progreso del asistente reiniciado.")
}

// --- helpers ------------------------------------------------------------

func (a *App) emitStepStatus(stepID, status string) {
	wruntime.EventsEmit(a.ctx, "step:status", map[string]string{"id": stepID, "status": status})
}

// elevatedRequestFor returns the (placeholder) elevated install command for a
// step, OS-aware. CP1 only needs these for Preview; the real install command
// strings are finalised in CP3.
func (a *App) elevatedRequestFor(id wizard.StepID) (elevate.Request, bool) {
	switch id {
	case wizard.StepVirtualBox:
		switch runtime.GOOS {
		case "windows":
			return elevate.Request{Command: "winget", Args: []string{"install", "-e", "--id", "Oracle.VirtualBox", "--accept-package-agreements", "--accept-source-agreements"}, Reason: "Instalar VirtualBox a nivel de sistema."}, true
		case "darwin":
			return elevate.Request{Command: "brew", Args: []string{"install", "--cask", "virtualbox"}, Reason: "Instalar VirtualBox (requiere aprobar la extensión de kernel)."}, true
		case "linux":
			return elevate.Request{Command: "apt-get", Args: []string{"install", "-y", "virtualbox"}, Reason: "Instalar VirtualBox desde los repositorios."}, true
		}
	case wizard.StepVagrant:
		switch runtime.GOOS {
		case "windows":
			return elevate.Request{Command: "winget", Args: []string{"install", "-e", "--id", "Hashicorp.Vagrant", "--accept-package-agreements", "--accept-source-agreements"}, Reason: "Instalar Vagrant a nivel de sistema."}, true
		case "darwin":
			return elevate.Request{Command: "brew", Args: []string{"install", "--cask", "vagrant"}, Reason: "Instalar Vagrant."}, true
		case "linux":
			return elevate.Request{Command: "apt-get", Args: []string{"install", "-y", "vagrant"}, Reason: "Instalar Vagrant desde los repositorios."}, true
		}
	}
	return elevate.Request{}, false
}

// elevationProbe returns a harmless command that, when run elevated, proves
// we obtained administrator/root on this OS.
func elevationProbe() elevate.Request {
	switch runtime.GOOS {
	case "windows":
		// whoami /groups includes the High Mandatory Level SID when elevated.
		return elevate.Request{Command: "whoami", Args: []string{"/groups"}, Reason: "Probar que la elevación funciona."}
	default:
		// macOS/Linux: under elevation, whoami prints "root".
		return elevate.Request{Command: "whoami", Reason: "Probar que la elevación funciona."}
	}
}

// probeConfirmsElevation interprets the probe output per OS.
func probeConfirmsElevation(stdout string) bool {
	s := strings.ToLower(stdout)
	if runtime.GOOS == "windows" {
		// S-1-16-12288 = High Mandatory Level (present only when elevated).
		return strings.Contains(s, "s-1-16-12288")
	}
	return strings.Contains(s, "root")
}

func stateFilePath() string {
	dir := executableDir()
	return filepath.Join(dir, ".vop_state.json")
}

func executableDir() string {
	exe, err := os.Executable()
	if err == nil {
		if resolved, err := filepath.EvalSymlinks(exe); err == nil {
			exe = resolved
		}
		return filepath.Dir(exe)
	}
	if cwd, err := os.Getwd(); err == nil {
		return cwd
	}
	return "."
}
