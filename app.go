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
	"github.com/abxda/vagrant-onboarding-panel/internal/labexercise"
	"github.com/abxda/vagrant-onboarding-panel/internal/logsink"
	"github.com/abxda/vagrant-onboarding-panel/internal/runner"
	"github.com/abxda/vagrant-onboarding-panel/internal/state"
	"github.com/abxda/vagrant-onboarding-panel/internal/tools"
	"github.com/abxda/vagrant-onboarding-panel/internal/vagrant"
	"github.com/abxda/vagrant-onboarding-panel/internal/wizard"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the Wails-bound application object.
type App struct {
	ctx   context.Context
	sink  *logsink.Sink
	runr  *runner.Runner
	state *state.State
}

// NewApp constructs the App with its log sink and process runner.
func NewApp() *App {
	sink := logsink.New(4000)
	return &App{
		sink: sink,
		runr: runner.New(sink),
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

// ToolStatus reports whether a step's tool is installed, plus which package
// manager would install it. Bound for the frontend to render "ya instalado
// vX" vs an install button.
type ToolStatus struct {
	Step         string `json:"step"`
	Installed    bool   `json:"installed"`
	Version      string `json:"version"`
	Path         string `json:"path"`
	PkgManager   string `json:"pkgManager"`
	PkgAvailable bool   `json:"pkgAvailable"`
}

// GetToolStatus detects whether the tool for a step (virtualbox|vagrant) is
// installed and reports the available package manager. Read-only.
func (a *App) GetToolStatus(stepID string) ToolStatus {
	ctx, cancel := context.WithTimeout(a.ctx, 20*time.Second)
	defer cancel()
	pm, pmOK := tools.PackageManager()
	ts := ToolStatus{Step: stepID, PkgManager: pm, PkgAvailable: pmOK}
	var info tools.Info
	switch wizard.StepID(stepID) {
	case wizard.StepVirtualBox:
		info = tools.DetectVirtualBox(ctx)
	case wizard.StepVagrant:
		info = tools.DetectVagrant(ctx)
	default:
		return ts
	}
	ts.Installed = info.Installed
	ts.Version = info.Version
	ts.Path = info.Path
	return ts
}

// CheckStep re-checks a step's status. Diagnostico runs the real diagnostic;
// virtualbox/vagrant run real tool detection; the rest (box/up/servidores)
// return their persisted value (real checks arrive in CP4).
func (a *App) CheckStep(stepID string) string {
	id := wizard.StepID(stepID)
	if id == wizard.StepDiagnostico {
		return string(a.GetDiagnostics().Overall)
	}
	if id == wizard.StepVirtualBox || id == wizard.StepVagrant {
		ts := a.GetToolStatus(stepID)
		st := string(wizard.StatusError)
		if ts.Installed {
			st = string(wizard.StatusOK)
			a.sink.Emit("INFO", fmt.Sprintf("%s detectado: versión %s (%s)", toolDisplayName(id), ts.Version, ts.Path))
		} else {
			a.sink.Emit("WARN", toolDisplayName(id)+" no está instalado.")
		}
		a.state.SetStatus(stepID, st)
		a.emitStepStatus(stepID, st)
		return st
	}
	if id == wizard.StepBox || id == wizard.StepUp {
		c, err := a.vagrantClient()
		if err != nil {
			a.sink.Emit("WARN", err.Error())
			a.state.SetStatus(stepID, string(wizard.StatusError))
			a.emitStepStatus(stepID, string(wizard.StatusError))
			return string(wizard.StatusError)
		}
		st := string(wizard.StatusError)
		if id == wizard.StepBox {
			if added, _ := c.BoxAdded(a.ctx); added {
				st = string(wizard.StatusOK)
				a.sink.Emit("INFO", "La caja "+vagrant.BoxName+" está añadida.")
			} else {
				a.sink.Emit("WARN", "La caja "+vagrant.BoxName+" todavía no está añadida.")
			}
		} else {
			vmState, _ := c.State(a.ctx)
			if vmState == "running" {
				st = string(wizard.StatusOK)
				a.sink.Emit("INFO", "La VM está corriendo.")
			} else {
				a.sink.Emit("WARN", "La VM no está corriendo (estado: "+vmState+").")
			}
		}
		a.state.SetStatus(stepID, st)
		a.emitStepStatus(stepID, st)
		return st
	}
	a.sink.Emit("INFO", "Verificando paso: "+stepID+" (verificación real llega en el siguiente checkpoint)")
	st := a.state.Status(stepID)
	if st == "" {
		st = string(wizard.StatusUnknown)
	}
	return st
}

func toolDisplayName(id wizard.StepID) string {
	if id == wizard.StepVirtualBox {
		return "VirtualBox"
	}
	return "Vagrant"
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

	// Steps 2 & 3 are real installs of VirtualBox / Vagrant.
	if id == wizard.StepVirtualBox || id == wizard.StepVagrant {
		return a.installTool(id)
	}

	// Steps 4-6 are real Vagrant operations.
	switch id {
	case wizard.StepBox:
		return a.runAddBox()
	case wizard.StepUp:
		return a.runVagrantUp()
	case wizard.StepServidores:
		return a.runServicesAndExercise()
	}
	return ActionResult{OK: false, Message: "Paso desconocido: " + stepID}
}

// --- CP4: Vagrant steps -------------------------------------------------

// vagrantClient resolves the vagrant binary and working directory, creating
// the working dir if needed.
func (a *App) vagrantClient() (*vagrant.Client, error) {
	ctx, cancel := context.WithTimeout(a.ctx, 15*time.Second)
	defer cancel()
	info := tools.DetectVagrant(ctx)
	if !info.Installed {
		return nil, fmt.Errorf("Vagrant no está instalado; completa el paso 3 primero")
	}
	workdir := a.state.VagrantDir
	if workdir == "" {
		workdir = vagrant.DefaultWorkDir()
		a.state.SetVagrantDir(workdir)
	}
	if err := os.MkdirAll(workdir, 0o755); err != nil {
		return nil, fmt.Errorf("no pude crear el directorio de trabajo %s: %w", workdir, err)
	}
	return vagrant.NewClient(info.Path, workdir, a.runr), nil
}

// runAddBox materialises the exercise + Vagrantfile, then downloads the lab box.
func (a *App) runAddBox() ActionResult {
	id := wizard.StepBox
	a.markRunning(id)
	c, err := a.vagrantClient()
	if err != nil {
		a.failStep(id)
		a.sink.Emit("ERROR", err.Error())
		return ActionResult{Message: err.Error()}
	}

	// Materialise the embedded exercise next to the Vagrantfile.
	exDir := filepath.Join(c.WorkDir, "ejercicio_01")
	if written, err := labexercise.Materialize(exDir); err != nil {
		a.failStep(id)
		a.sink.Emit("ERROR", "No pude escribir el ejercicio: "+err.Error())
		return ActionResult{Message: "No pude preparar el ejercicio."}
	} else {
		a.sink.Emit("INFO", "Ejercicio_01 preparado: "+strings.Join(written, ", "))
	}
	if err := c.EnsureVagrantfile(); err != nil {
		a.failStep(id)
		a.sink.Emit("ERROR", "No pude generar el Vagrantfile: "+err.Error())
		return ActionResult{Message: "No pude generar el Vagrantfile."}
	}
	a.sink.Emit("INFO", "Vagrantfile generado en "+c.WorkDir)

	if added, _ := c.BoxAdded(a.ctx); added {
		a.sink.Emit("INFO", "La caja "+vagrant.BoxName+" ya estaba añadida.")
		a.okStep(id)
		return ActionResult{OK: true, Message: "La caja ya estaba añadida."}
	}

	a.sink.Emit("INFO", "Descargando la caja "+vagrant.BoxName+" (~4.4 GB). Puede tardar varios minutos según tu conexión…")
	res, err := c.AddBox(a.ctx)
	if err != nil {
		a.failStep(id)
		a.sink.Emit("ERROR", "Error en vagrant box add: "+err.Error())
		return ActionResult{Message: "Falló la descarga de la caja.", Detail: err.Error()}
	}
	if res.ExitCode != 0 {
		a.failStep(id)
		return ActionResult{Message: fmt.Sprintf("vagrant box add terminó con código %d. Revisa el registro.", res.ExitCode)}
	}
	if added, _ := c.BoxAdded(a.ctx); added {
		a.okStep(id)
		a.sink.Emit("INFO", "✓ Caja añadida correctamente.")
		return ActionResult{OK: true, Message: "Caja añadida correctamente."}
	}
	a.failStep(id)
	return ActionResult{Message: "La caja no aparece tras la descarga."}
}

// runVagrantUp boots the VM.
func (a *App) runVagrantUp() ActionResult {
	id := wizard.StepUp
	a.markRunning(id)
	c, err := a.vagrantClient()
	if err != nil {
		a.failStep(id)
		a.sink.Emit("ERROR", err.Error())
		return ActionResult{Message: err.Error()}
	}
	_ = c.EnsureVagrantfile()

	if st, _ := c.State(a.ctx); st == "running" {
		a.sink.Emit("INFO", "La VM ya está corriendo.")
		a.okStep(id)
		return ActionResult{OK: true, Message: "La VM ya estaba corriendo."}
	}

	a.sink.Emit("INFO", "Levantando la VM con 'vagrant up'. El primer arranque puede tardar varios minutos (VirtualBox correrá sobre Hyper-V, en modo compatibilidad)…")
	if _, err := c.Up(a.ctx); err != nil {
		a.failStep(id)
		a.sink.Emit("ERROR", "Error en vagrant up: "+err.Error())
		return ActionResult{Message: "Falló el arranque de la VM.", Detail: err.Error()}
	}
	st, _ := c.State(a.ctx)
	if st == "running" {
		a.okStep(id)
		a.sink.Emit("INFO", "✓ VM levantada y en ejecución.")
		return ActionResult{OK: true, Message: "VM levantada y corriendo."}
	}
	a.failStep(id)
	return ActionResult{Message: "La VM no quedó en estado 'running' (estado: " + st + "). Revisa el registro."}
}

// runServicesAndExercise uploads the exercise into the VM and runs the
// WordCount playbook step by step via `vagrant ssh -c`.
func (a *App) runServicesAndExercise() ActionResult {
	id := wizard.StepServidores
	a.markRunning(id)
	c, err := a.vagrantClient()
	if err != nil {
		a.failStep(id)
		a.sink.Emit("ERROR", err.Error())
		return ActionResult{Message: err.Error()}
	}

	if st, _ := c.State(a.ctx); st != "running" {
		a.failStep(id)
		msg := "La VM no está corriendo (estado: " + st + "). Ejecuta el paso 5 primero."
		a.sink.Emit("ERROR", msg)
		return ActionResult{Message: msg}
	}

	// Upload the exercise into the VM (SCP — no Guest Additions needed).
	exDir := filepath.Join(c.WorkDir, "ejercicio_01")
	a.sink.Emit("INFO", "Subiendo el Ejercicio_01 a la VM ("+labexercise.RemoteDir+")…")
	if res, err := c.Upload(a.ctx, exDir, labexercise.RemoteDir); err != nil || res.ExitCode != 0 {
		a.sink.Emit("WARN", "No pude subir el ejercicio con 'vagrant upload'; continuaré por si ya estaba en la VM.")
	}

	// Run the WordCount playbook.
	steps := labexercise.Steps()
	a.sink.Emit("INFO", "Ejecutando el Ejercicio_01 (WordCount) dentro de la VM, paso a paso:")
	for i, s := range steps {
		a.sink.Emit("INFO", strings.Repeat("─", 56))
		a.sink.Emit("INFO", fmt.Sprintf("[%d/%d] %s", i+1, len(steps), s.Title))
		if s.Notes != "" {
			a.sink.Emit("INFO", "  ↳ "+s.Notes)
		}
		a.sink.Emit("INFO", "$ "+s.Cmd)
		res, err := c.SSH(a.ctx, 20*time.Minute, s.Cmd)
		if err != nil {
			a.failStep(id)
			a.sink.Emit("ERROR", "Error: "+err.Error())
			return ActionResult{Message: fmt.Sprintf("El paso %d del ejercicio falló: %v", i+1, err)}
		}
		if res.ExitCode != 0 {
			a.failStep(id)
			return ActionResult{Message: fmt.Sprintf("El paso %d del ejercicio (%s) terminó con código %d. Revisa el registro.", i+1, s.Title, res.ExitCode)}
		}
	}
	a.okStep(id)
	a.sink.Emit("INFO", "✓ Ejercicio_01 (WordCount) completado dentro de la VM.")
	return ActionResult{OK: true, Message: "Servicios verificados y Ejercicio_01 completado en la VM."}
}

func (a *App) markRunning(id wizard.StepID) {
	a.state.SetStatus(string(id), string(wizard.StatusRunning))
	a.emitStepStatus(string(id), string(wizard.StatusRunning))
}

func (a *App) okStep(id wizard.StepID) {
	a.state.SetStatus(string(id), string(wizard.StatusOK))
	a.emitStepStatus(string(id), string(wizard.StatusOK))
}

// installTool performs the real detect → (maybe) elevated install → re-detect
// flow for VirtualBox or Vagrant.
func (a *App) installTool(id wizard.StepID) ActionResult {
	name := toolDisplayName(id)
	a.state.SetStatus(string(id), string(wizard.StatusRunning))
	a.emitStepStatus(string(id), string(wizard.StatusRunning))

	// 1. Already installed? Then we're done — never reinstall behind the user's back.
	if ts := a.GetToolStatus(string(id)); ts.Installed {
		a.sink.Emit("INFO", fmt.Sprintf("✓ %s ya está instalado: versión %s", name, ts.Version))
		a.state.SetStatus(string(id), string(wizard.StatusOK))
		a.emitStepStatus(string(id), string(wizard.StatusOK))
		return ActionResult{OK: true, Message: name + " ya estaba instalado (v" + ts.Version + ")."}
	}

	// 2. Need a package manager.
	pm, pmOK := tools.PackageManager()
	if !pmOK {
		a.failStep(id)
		msg := "No encontré un gestor de paquetes para instalar automáticamente. Instala " + name + " manualmente desde su sitio oficial."
		a.sink.Emit("ERROR", msg)
		return ActionResult{OK: false, Message: msg}
	}

	req, ok := a.elevatedRequestFor(id)
	if !ok {
		a.failStep(id)
		return ActionResult{OK: false, Message: "No hay un comando de instalación definido para esta plataforma."}
	}

	// 3. Run the install elevated, showing the exact command first.
	a.sink.Emit("INFO", fmt.Sprintf("Instalando %s con %s. Se pedirá aprobación de administrador (%s).", name, pm, elevate.Mechanism()))
	a.sink.Emit("INFO", "Comando exacto (elevado): "+elevate.Preview(req))
	a.sink.Emit("INFO", "Aprueba el diálogo del sistema. La descarga puede tardar varios minutos…")

	ctx, cancel := context.WithTimeout(a.ctx, 25*time.Minute)
	defer cancel()
	res, err := elevate.Run(ctx, req)
	for _, ln := range sanitizeInstallerOutput(res.Stdout) {
		a.sink.Emit("INFO", ln)
	}
	for _, ln := range sanitizeInstallerOutput(res.Stderr) {
		a.sink.Emit("WARN", ln)
	}
	if err != nil {
		a.failStep(id)
		a.sink.Emit("ERROR", "Error al instalar: "+err.Error())
		return ActionResult{OK: false, Message: "Falló la instalación de " + name + ".", Detail: err.Error()}
	}
	if res.Cancelled {
		a.failStep(id)
		a.sink.Emit("WARN", "Cancelaste el diálogo de elevación; no se instaló nada.")
		return ActionResult{OK: false, Cancelled: true, Message: "Cancelaste la instalación de " + name + "."}
	}

	// 4. Re-detect to confirm the install actually landed.
	if ts := a.GetToolStatus(string(id)); ts.Installed {
		a.sink.Emit("INFO", fmt.Sprintf("✓ %s instalado correctamente: versión %s", name, ts.Version))
		a.state.SetStatus(string(id), string(wizard.StatusOK))
		a.emitStepStatus(string(id), string(wizard.StatusOK))
		return ActionResult{OK: true, Message: name + " instalado (v" + ts.Version + ")."}
	}

	// Installer reported success but we can't find the binary yet. Common with
	// winget when PATH/env hasn't refreshed — tell the user how to verify.
	a.failStep(id)
	a.sink.Emit("WARN", fmt.Sprintf("El instalador terminó (código %d) pero todavía no detecto %s. Pulsa 'Verificar estado'; si sigue sin aparecer, reinicia el panel.", res.ExitCode, name))
	return ActionResult{OK: false, Message: name + " no se detectó tras la instalación. Pulsa 'Verificar estado' o reinicia el panel."}
}

func (a *App) failStep(id wizard.StepID) {
	a.state.SetStatus(string(id), string(wizard.StatusError))
	a.emitStepStatus(string(id), string(wizard.StatusError))
}

// sanitizeInstallerOutput strips the visual noise package managers like winget
// emit (animated progress bars and spinner frames) so the live log — and the
// "Copiar consola" report a student sends to the teacher — stays readable.
// We keep the meaningful lines (Found …, Downloading …, Successfully installed).
func sanitizeInstallerOutput(raw string) []string {
	if raw == "" {
		return nil
	}
	var out []string
	for _, ln := range strings.Split(raw, "\n") {
		t := strings.TrimSpace(strings.TrimRight(ln, "\r"))
		if t == "" {
			continue
		}
		// Drop progress-bar lines (block-drawing chars) entirely; the
		// "Downloading <url>" line already tells the student what's happening.
		if strings.ContainsAny(t, "▒█░▓") {
			continue
		}
		// Drop spinner-only frames ("- \ | /" possibly with surrounding spaces).
		if isSpinnerLine(t) {
			continue
		}
		out = append(out, t)
	}
	return out
}

func isSpinnerLine(s string) bool {
	for _, r := range s {
		switch r {
		case '-', '\\', '|', '/', ' ', '\t':
			// spinner glyph or whitespace; keep scanning
		default:
			return false
		}
	}
	return true
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
