# AGENTS.md — Compilar el Panel de Onboarding (Edición Vagrant)

Guía para agentes (Linux x86-64, macOS Intel) que compilan el **Panel de
Onboarding · Edición Vagrant** de la suite *Big Data Lab* del Dr. Abel
Coronado. Esta edición aplica a **Linux x86-64** y **macOS Intel** (no Apple
Silicon, que usa la edición Portable).

## Qué es

App de escritorio (Wails: Go + HTML/CSS/JS) que guía al alumno a montar un
laboratorio de Big Data sobre **Vagrant + VirtualBox** (caja
`abxda/big-data-lab`), y luego le da un espacio de trabajo (servicios,
ejercicio WordCount, explorador HDFS, Jupyter).

Arquitectura de la UI (importante para mantener paridad con el Portable):
- **Preparación** (capa de despliegue, propia de esta edición): los pasos
  diagnóstico → VirtualBox → Vagrant → caja → levantar VM.
- **Mi Laboratorio** (capa de trabajo, IDÉNTICA al Portable): Servicios,
  Ejercicios, HDFS, Jupyter/Archivos. No reescribas esta capa: debe quedar
  homóloga al `bdpv6-launcher`.

## Requisitos de compilación

- Go 1.23+
- Wails v2.12.0: `go install github.com/wailsapp/wails/v2/cmd/wails@v2.12.0`
- CGO habilitado (Wails lo necesita) + toolchain del SO:
  - **Linux**: `gcc`, `pkg-config`, `libgtk-3-dev`, `libwebkit2gtk-4.0-dev`
  - **macOS Intel**: Xcode Command Line Tools (`xcode-select --install`)
- `wails doctor` debe decir "Your system is ready".

## Compilar

```bash
# Linux x86-64
wails build -platform linux/amd64
# -> build/bin/vagrant-onboarding-panel

# macOS Intel
wails build -platform darwin/amd64
# -> build/bin/vagrant-onboarding-panel.app
```

## Verificar (sin VirtualBox/Vagrant instalados)

La app arranca y muestra el diagnóstico aunque no haya VirtualBox/Vagrant —
ese es justo el primer paso. Comprueba:

1. La app abre y muestra el sidebar con **Preparación** + **Mi laboratorio**.
2. Paso 1 (Diagnóstico): detecta VT-x/AMD-V (Linux `/proc/cpuinfo` vmx/svm;
   macOS `sysctl machdep.cpu.features` VMX), RAM y disco.
3. La elevación nativa funciona: botón de instalar VirtualBox debería disparar
   `pkexec` (Linux) o el diálogo de admin de macOS (`osascript`).

Prueba end-to-end completa (descargar caja 4.4 GB + `vagrant up` + ejercicio)
solo si tienes VirtualBox + Vagrant y ~15 GB libres. El WordCount ya está
validado en Windows; en Linux/Mac Intel debería comportarse igual (la caja es
Debian 11 amd64).

## Diferencias por plataforma ya codificadas (NO reimplementar)

| Tema | Cómo está resuelto |
|------|--------------------|
| Elevación | `internal/elevate`: pkexec (Linux), osascript (macOS) |
| Gestor de paquetes | apt/dnf/pacman (Linux), brew (macOS) — `internal/tools` |
| Consola SSH | gnome-terminal/xterm (Linux), Terminal.app (macOS) — `internal/vagrant/ssh_unix.go` |
| Abrir carpeta/URL | xdg-open (Linux), open (macOS) — `internal/desktop` |
| Diagnóstico VT-x | `internal/diagnose/diagnose_{linux,darwin}.go` |

> En Linux/macOS NO hay conflicto de hipervisor como el Hyper-V de Windows, así
> que la VM corre a velocidad nativa (sin "tortuga").

## Publicar

Binarios a **GitHub Releases** de este repo. El binario Wails es per-plataforma
(no universal), así que cada SO sube el suyo.

## Token de Hugging Face

Para compilar y verificar **NO** necesitas token. El token de HF solo lo usa
el Meta-Launcher para publicar distros, y nunca va en git (te lo pasa el Dr.
Coronado por fuera; ver el AGENTS.md de `bdp-meta-launcher`).

## Reportar de vuelta

Issue o PR con: `uname -a`, captura/arranque OK, salida del diagnóstico, y el
binario subido al Release. Si algo de la capa **Mi Laboratorio** se ve distinto
al Portable, márcalo — debe ser homólogo.

Autoría: **Dr. Abel Coronado**.
