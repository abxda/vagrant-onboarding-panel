// Package labexercise embeds the WordCount exercise (mapper.py, reducer.py,
// breweries.csv) and defines the step-by-step playbook to run it INSIDE the
// Vagrant VM via `vagrant ssh -c`. This mirrors the portable launcher's
// Ejercicio_01 so students get the same pedagogical flow on Plan B.
//
// Key difference vs the portable (Windows) version: inside the Debian 11 VM
// the interpreter is `python3` (Debian 11 ships no bare `python`), and paths
// are clean POSIX paths, so no MSYS path-mangling workarounds are needed.
package labexercise

import (
	"bytes"
	"embed"
	"os"
	"path/filepath"
)

//go:embed files/ejercicio_01/*
var files embed.FS

// RemoteDir is where the exercise files live inside the VM (a Vagrant synced
// folder maps the host materialised dir here).
const RemoteDir = "/home/vagrant/ejercicio_01"

// Materialize writes the embedded exercise files into destDir on the host so
// Vagrant can mount them into the VM as a synced folder. Returns the list of
// written filenames.
func Materialize(destDir string) ([]string, error) {
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return nil, err
	}
	entries, err := files.ReadDir("files/ejercicio_01")
	if err != nil {
		return nil, err
	}
	var written []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		data, err := files.ReadFile("files/ejercicio_01/" + e.Name())
		if err != nil {
			return nil, err
		}
		// Normalize CRLF → LF. These files run inside a Debian 11 VM; a stray
		// carriage return would corrupt the CSV tokens and could break the
		// Python scripts. This guarantees LF no matter how the repo was
		// checked out (autocrlf, Windows clone, etc.).
		data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
		if err := os.WriteFile(filepath.Join(destDir, e.Name()), data, 0o644); err != nil {
			return nil, err
		}
		written = append(written, e.Name())
	}
	return written, nil
}

// Step is one command of the WordCount playbook, run inside the VM.
type Step struct {
	Title string // shown as a header in the log
	Notes string // teacher's note (why)
	Cmd   string // the shell command executed via `vagrant ssh -c`
}

// StartServicesCmd starts the box's Big Data stack (HDFS, Kafka, Elasticsearch,
// Jupyter). Run with sudo (passwordless in the box).
//
// We do NOT call quasar-start.sh blindly. Real failure seen in testing: a
// half-dead Kafka left a zombie JVM holding the controller port 9093 →
// "Address already in use" → Kafka never came back, and quasar-start.sh
// doesn't clean zombies. So we:
//  1. Kill any Kafka zombie that is NOT a healthy broker (holds 9093 but not
//     9092), freeing the controller port. Healthy Kafka is left alone.
//  2. Run quasar-start.sh (idempotent for the other services).
//
// CRITICAL (dos motivos):
//  1. stdio redirigido a un archivo (</dev/null >log 2>&1): si no, el canal SSH
//     nunca cierra — Elasticsearch (-d) y Jupyter (&) heredan el stdout del canal
//     y `vagrant ssh -c` se cuelga hasta el timeout.
//  2. setsid: arranca quasar-start.sh en una SESIÓN NUEVA, desacoplada de la
//     sesión SSH. SIN ESTO, al cerrarse la sesión los daemons reciben SIGHUP y
//     mueren ~1s después (se observó: NameNode/DataNode/Elasticsearch caían
//     justo al volver el comando; solo Kafka sobrevivía). Con setsid el stack
//     queda en su propio grupo de sesión y sigue vivo.
// Borramos el log previo (puede ser de root de un arranque anterior) para que el
// redirect no falle por permisos.
const StartServicesCmd = `bash -lc '
# Si Kafka tiene el controller (9093) pero NO el broker (9092), es un zombie:
# lo matamos para liberar el puerto y que arranque limpio.
if ss -tln 2>/dev/null | grep -q ":9093 " && ! ss -tln 2>/dev/null | grep -q ":9092 "; then
  echo "Kafka quedó a medias (zombie en 9093); lo reinicio limpio…"
  sudo pkill -9 -f kafka 2>/dev/null || true
  sleep 3
fi
sudo rm -f /tmp/quasar-start.log
sudo setsid bash -lc "/usr/local/bin/quasar-start.sh > /tmp/quasar-start.log 2>&1" </dev/null
echo "Stack de servicios lanzado en segundo plano (setsid); detalle en /tmp/quasar-start.log dentro de la VM."
' </dev/null >/tmp/quasar-wrap.log 2>&1; cat /tmp/quasar-wrap.log`

// WaitHDFSCmd polls until the HDFS NameNode RPC answers (or ~90s elapse), so we
// don't run the job before the daemons finished booting.
const WaitHDFSCmd = `for i in $(seq 1 45); do hdfs dfsadmin -safemode get >/dev/null 2>&1 && { echo "HDFS responde."; exit 0; }; echo "esperando a HDFS... ($i)"; sleep 2; done; echo "HDFS no respondió a tiempo"; exit 1`

// StopServicesCmd stops the stack (used by the SSH console help / future halt).
const StopServicesCmd = "sudo /usr/local/bin/quasar-stop.sh"

// Steps returns the ordered WordCount playbook for execution inside the VM.
// Every command runs as the `vagrant` user over SSH.
//
//   - hadoopOpts is prepended to step 5 to make Hadoop 3.3.6 tolerate the
//     box's OpenJDK 17 (the --add-opens flags); harmless if not needed.
func Steps() []Step {
	const hdfsInput = "/ej01/input"
	const hdfsOutput = "/ej01/output"
	// --add-opens flags make Hadoop 3.3.6 run under JDK 17 (the box's JDK).
	// LocalJobRunner is the least likely to need them, but they don't hurt.
	addOpens := `export HADOOP_OPTS="$HADOOP_OPTS ` +
		`--add-opens java.base/java.lang=ALL-UNNAMED ` +
		`--add-opens java.base/java.util=ALL-UNNAMED ` +
		`--add-opens java.base/java.lang.reflect=ALL-UNNAMED ` +
		`--add-opens java.base/java.nio=ALL-UNNAMED ` +
		`--add-opens java.base/sun.nio.ch=ALL-UNNAMED"`

	return []Step{
		{
			Title: "1 · Esperar a que HDFS salga de safe mode",
			Notes: "Al arrancar, el NameNode entra en 'safe mode' mientras carga su estado. Este comando espera a que termine; si ya está listo, vuelve al instante.",
			Cmd:   "hdfs dfsadmin -safemode wait",
		},
		{
			Title: "2 · Preparar directorio de entrada en HDFS",
			Notes: "Crea " + hdfsInput + " en HDFS si no existe. -p evita error si ya estaba.",
			Cmd:   "hdfs dfs -mkdir -p " + hdfsInput,
		},
		{
			Title: "3 · Subir el dataset (breweries.csv) a HDFS",
			Notes: "Copia el CSV local (montado en la VM) al directorio de entrada. -f sobrescribe si ya estaba.",
			Cmd:   "hdfs dfs -put -f " + RemoteDir + "/breweries.csv " + hdfsInput + "/",
		},
		{
			Title: "4 · Borrar salida anterior (idempotente)",
			Notes: "Hadoop se niega a escribir sobre una carpeta de salida existente; la borramos por si corres el job más de una vez.",
			Cmd:   "hdfs dfs -rm -r -f -skipTrash " + hdfsOutput + " 2>/dev/null; true",
		},
		{
			Title: "5 · Ejecutar el job MapReduce vía Hadoop Streaming",
			Notes: "Hadoop Streaming pasa cada línea al stdin del mapper y su stdout al reducer. Forzamos LocalJobRunner (sin YARN) y usamos python3 (Debian 11 no tiene 'python' a secas).",
			Cmd: addOpens + "; " +
				"hadoop jar $(ls /opt/bdpv5/hadoop/share/hadoop/tools/lib/hadoop-streaming-*.jar 2>/dev/null | head -1) " +
				"-D mapreduce.framework.name=local " +
				"-D mapreduce.jobtracker.address=local " +
				"-mapper 'python3 " + RemoteDir + "/mapper.py' " +
				"-reducer 'python3 " + RemoteDir + "/reducer.py' " +
				"-input " + hdfsInput + " -output " + hdfsOutput,
		},
		{
			Title: "6 · Top-20 palabras más frecuentes",
			Notes: "Leemos la salida del reducer (part-00000), ordenamos por la columna 2 numérica descendente y mostramos los 20 primeros tokens.",
			Cmd:   "hdfs dfs -cat " + hdfsOutput + "/part-00000 | sort -t$'\\t' -k2 -nr | head -20",
		},
	}
}
