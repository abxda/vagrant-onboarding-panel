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

## Estado v0.1.2 — qué validar (aprendido validando en Windows)

El panel ya trae los arreglos v0.1.2 (validados en Windows). En Linux deben
comportarse igual: el guest es el mismo Debian 11 amd64 y el código host es
multiplataforma. Confirma:

- **"Levantar todos los servicios"** arranca el stack con `setsid` DENTRO del
  guest: los daemons (HDFS/Elasticsearch/Jupyter) **siguen vivos al cerrarse la
  sesión SSH** (antes morían por SIGHUP — solo Kafka sobrevivía). Tras pulsarlo,
  HDFS debe quedar **activo y persistir**. La verificación reintenta ~100 s
  (HDFS y ES tardan en subir).
- **"Mi laboratorio"** se desbloquea cuando la VM está EN EJECUCIÓN (detección
  por NOMBRE con `VBoxManage`, sin depender del frágil `vagrant status`).
- Al **cerrar el panel** apaga la VM limpiamente (overlay "Cerrando de forma
  segura" → detiene servicios → ACPI por nombre → espera apagado).
- **Logs claros** (versión, rutas, nombre de la VM, cada decisión de estado):
  los alumnos están remotos, nada sin loguear.
- **VirtualBox en Linux**: tras instalarlo puede requerir cargar el módulo de
  kernel `vboxdrv` (Secure Boot puede bloquearlo). Valida que `vagrant up`
  levante con el provider `virtualbox`.

## Publicar (deploy a Hugging Face)

> Runbook completo (comandos, verificación de `oid`, ritual anti-clobber, costo de
> cada cambio): **`Curso_BDP/PUBLICAR_Y_ACTUALIZAR.md` §4** (panel chico ~5 MB).
> Este panel **no** tiene auto-update: el meta-launcher baja el tarball y verifica
> el SHA del manifiesto, así que aquí **sí** se edita `manifest.txt`.

El Meta-Launcher **descarga el panel desde el dataset HF `abxda/bdp-lab`** (no
desde Releases). Por eso hay que **empaquetar el panel en un `.tar.gz`, subirlo a
HF y añadir su entrada al `manifest.txt`**:

- **Raíz del tar**: el binario `vagrant-onboarding-panel` (Linux) o el bundle
  `vagrant-onboarding-panel.app` (macOS Intel) + `README.md`.
- **Nombre por convención**: `bdp-vagrant-<os>-amd64.tar.gz`
  (ej. `bdp-vagrant-linux-amd64.tar.gz`).
- **Clave del manifest**: `<os>-amd64-vagrant` (ej. `linux-amd64-vagrant`), con
  `.file`, `.sha256`, `.launch` (el binario/app) y `.size`.
- **Edición SIEMPRE aditiva y POR CLAVE**: baja el manifest actual, añade/
  actualiza SOLO tu clave, **no toques las entradas de otras plataformas**
  (candado anti-clobber). Sube el tar PRIMERO, el manifest DESPUÉS, y verifica
  que el sha publicado coincide con el del tar.
- (Opcional) copia del binario en GitHub Releases.

## Token de Hugging Face

Para **compilar/verificar NO** necesitas token. Para **publicar el tar a HF SÍ**
(rol *write*). El Dr. Coronado te deja el token en un archivo dentro de tu
carpeta de trabajo (p.ej. `.hf_token`, una sola línea, sin espacios). **Nunca**
va en git ni en logs.

## Reportar de vuelta

Issue o PR con: `uname -a`, captura/arranque OK, salida del diagnóstico, y el
binario subido al Release. Si algo de la capa **Mi Laboratorio** se ve distinto
al Portable, márcalo — debe ser homólogo.

Autoría: **Dr. Abel Coronado**.
