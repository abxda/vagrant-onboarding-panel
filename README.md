# Panel de Onboarding — Big Data Lab (Plan B, Vagrant)

Aplicación de escritorio multiplataforma (Wails: Go + web) que guía a estudiantes
para levantar un laboratorio de Big Data sobre **Vagrant + VirtualBox**, usando la
caja preconstruida [`abxda/big-data-lab`](https://portal.cloud.hashicorp.com/vagrant/discover/abxda/big-data-lab)
(Debian 11 con Hadoop 3.3.6, Kafka 4.0 KRaft, Elasticsearch 8.14.1, OpenJDK 17,
Python 3.9 + Jupyter Lab).

Es el **Plan B** del laboratorio: para estudiantes a quienes la distribución
*portable* no les funciona. Replica el mismo flujo pedagógico (incluido el
Ejercicio_01 WordCount con Hadoop Streaming), pero dentro de una máquina virtual.

> Autoría: **Dr. Abel Coronado**.

## Principio de diseño: elevación puntual, nunca app-como-admin

La aplicación **siempre corre sin privilegios**. Cuando un paso necesita
administrador (instalar VirtualBox o Vagrant), eleva **solo ese comando** con el
mecanismo nativo del sistema, mostrándote antes el comando exacto que se ejecutará:

| SO            | Mecanismo de elevación                              |
|---------------|-----------------------------------------------------|
| Windows       | `ShellExecuteEx` con verbo `runas` → diálogo UAC    |
| macOS (Intel) | `osascript … with administrator privileges`         |
| Linux         | `pkexec` (PolicyKit, diálogo gráfico)               |

## Plataformas objetivo (solo x86-64)

- Windows x64
- Linux x64
- macOS **Intel** x64 (no Apple Silicon)

Provider de virtualización: **siempre VirtualBox**.

## El asistente (6 pasos)

1. **Diagnóstico** — VT-x/AMD-V, RAM y disco, conflictos de hipervisor (Hyper-V/WSL2/Integridad de memoria). Solo lectura.
2. **VirtualBox** — detectar/instalar (winget / brew --cask / repos). En macOS Intel: aprobar la extensión de kernel (manual, guiado).
3. **Vagrant** — detectar/instalar.
4. **Caja** — `vagrant box add` de `abxda/big-data-lab`.
5. **Levantar** — generar Vagrantfile + `vagrant up` (salida `--machine-readable`).
6. **Servicios + ejercicio** — iniciar HDFS y dejar listo el Ejercicio_01, vía `vagrant ssh -c`.

## Matriz de compatibilidad y limitaciones conocidas

| Tema | Estado | Nota |
|------|--------|------|
| Hadoop 3.3.6 en OpenJDK 17 | ⚠️ no soportado oficialmente | LocalJobRunner (sin YARN) es el caso menos propenso a fallar; si rompe, se aplican flags `--add-opens` o se usa JDK 11 para Hadoop. Se valida al levantar la VM. |
| `python mapper.py` en Debian 11 | ⚠️ requiere ajuste | Debian 11 no trae el symlink `python`, solo `python3`. El ejercicio usa `python3` explícito. |
| Kafka 4.0 KRaft | ✅ | Requiere JDK 17 (ok en la caja). Usa `--bootstrap-server`, no `--zookeeper`. |
| Elasticsearch 8.14 | ✅ | Seguridad ON por defecto (HTTPS + password autogenerada). |
| VirtualBox en Windows 11 con Hyper-V/VBS | ✅ convive | Estrategia elegida: NO desactivamos el hipervisor ni la seguridad del alumno. VirtualBox 7 corre sobre Hyper-V (modo compatibilidad, ícono de tortuga, más lento pero funcional). El Diagnóstico lo detecta e informa sin alarmar. Respetuoso con equipos administrados/corporativos donde `bcdedit` está bloqueado. |
| macOS Intel: kext de VirtualBox | ⚠️ manual | La aprobación de la extensión de kernel NO se automatiza; el panel da instrucciones visuales. |
| Apple Silicon | ❌ fuera de alcance | VirtualBox + esta caja amd64 no aplican. |

## Compilar

Requiere Go 1.23+, Wails v2.12, y (Windows) gcc de MSYS2 con `CGO_ENABLED=1`.

```bash
# Windows
wails build -platform windows/amd64

# Linux
wails build -platform linux/amd64

# macOS Intel
wails build -platform darwin/amd64
```

El binario queda en `build/bin/`.

## Cómo enviar resultados al profesor

Cada paso tiene un panel **Registro en vivo** con un botón **"Copiar consola"**.
El alumno lo pulsa y pega el contenido en su mensaje: el texto incluye un
encabezado con su sistema operativo, arquitectura, versión del panel y el paso
en el que está, además de todo el log. Así el profesor (Dr. Abel Coronado)
recibe el contexto completo para diagnosticar sin pedir más datos.

## Estado del proyecto (checkpoints)

- [x] **CP1** — Esqueleto Wails + wizard (6 pasos) + módulo de elevación nativa.
      Probado en Windows: el botón "Probar elevación" dispara UAC y confirma admin.
- [x] **CP2** — Paso de Diagnóstico real (VT-x vía WMI, RAM/disco vía gopsutil,
      hipervisor/VBS, Integridad de memoria). Estrategia respetuosa: convivimos
      con Hyper-V (VirtualBox en modo compatibilidad), nunca desactivamos la
      seguridad del alumno.
- [x] **CP3** — Detección e instalación real de VirtualBox y Vagrant
      (winget/brew/apt, elevado, mostrando el comando exacto; re-detecta para
      confirmar). Si ya están instalados, no reinstala.
- [x] **Extra** — Botón "Copiar consola" con encabezado de contexto para soporte.
- [ ] **CP4** — Añadir la caja `abxda/big-data-lab`, `vagrant up` (salida
      `--machine-readable`), iniciar servicios y dejar listo el Ejercicio_01
      (WordCount con Hadoop Streaming) ejecutable paso a paso vía `vagrant ssh -c`.

### Notas de prueba en máquina real (Windows 11)

- Equipo de prueba: 31.4 GB RAM, 298 GB libres, VT-x soportada, **VBS/Integridad
  de memoria activas** (Hyper-V presente) → el diagnóstico lo reporta como aviso
  educativo, no como bloqueo. VirtualBox correrá en modo compatibilidad.
- VirtualBox y Vagrant **no** estaban instalados (escenario real del alumno):
  los pasos 2 y 3 los instalan vía `winget` con aprobación UAC.
