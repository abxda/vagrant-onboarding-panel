// Package wizard defines the sequential onboarding steps and tracks their
// completion state. The actual work each step performs is implemented in
// app.go (which has access to the runner, elevate and diagnostics packages);
// this package owns the step METADATA and the persisted progress so the UI
// can render the wizard and resume where the student left off.
package wizard

// StepID identifies a wizard step.
type StepID string

const (
	StepDiagnostico StepID = "diagnostico"
	StepVirtualBox  StepID = "virtualbox"
	StepVagrant     StepID = "vagrant"
	StepBox         StepID = "box"
	StepUp          StepID = "up"
	StepServidores  StepID = "servidores"
)

// Status is a step's traffic-light state.
type Status string

const (
	StatusUnknown Status = "unknown" // not checked yet
	StatusOK      Status = "ok"      // verified good (green)
	StatusWarn    Status = "warn"    // works but with caveats (yellow)
	StatusError   Status = "error"   // blocked / failed (red)
	StatusRunning Status = "running" // an action is executing
)

// Step is the static definition of one wizard step.
type Step struct {
	ID    StepID `json:"id"`
	Index int    `json:"index"` // 1-based display order
	Title string `json:"title"`
	// Why explains the pedagogical purpose (what + why), shown in the card.
	Why string `json:"why"`
	// NeedsElevation is true if this step's primary action requires admin/root.
	NeedsElevation bool `json:"needsElevation"`
	// ActionLabel is the verb on the primary button ("Instalar", "Verificar"…).
	ActionLabel string `json:"actionLabel"`
}

// Steps returns the ordered list of wizard steps with Spanish copy.
func Steps() []Step {
	return []Step{
		{
			ID:             StepDiagnostico,
			Index:          1,
			Title:          "Diagnóstico del sistema",
			Why:            "Antes de instalar nada, comprobamos que tu equipo puede correr máquinas virtuales: virtualización por hardware (VT-x/AMD-V), RAM y disco suficientes, y si Windows tiene activo su hipervisor de seguridad (Hyper-V / VBS / Integridad de memoria). Importante: si está activo, NO lo desactivamos ni tocamos tu seguridad — VirtualBox correrá por encima de él en modo compatibilidad (un poco más lento, con un ícono de tortuga: es normal). Este paso es de SOLO LECTURA: no cambia absolutamente nada en tu equipo.",
			NeedsElevation: false,
			ActionLabel:    "Ejecutar diagnóstico",
		},
		{
			ID:             StepVirtualBox,
			Index:          2,
			Title:          "Instalar VirtualBox",
			Why:            "VirtualBox es el motor de virtualización que ejecutará la máquina con todo el stack de Big Data ya instalado. Detectamos si ya lo tienes; si no, lo instalamos con el gestor de paquetes nativo de tu sistema. La instalación necesita privilegios de administrador, por eso se te pedirá aprobación del sistema en ese momento.",
			NeedsElevation: true,
			ActionLabel:    "Instalar VirtualBox",
		},
		{
			ID:             StepVagrant,
			Index:          3,
			Title:          "Instalar Vagrant",
			Why:            "Vagrant automatiza la creación y arranque de la máquina virtual a partir de una 'caja' (imagen) preconstruida. Así evitas configurar Hadoop, Kafka y Elasticsearch a mano. Igual que VirtualBox, su instalación requiere privilegios de administrador.",
			NeedsElevation: true,
			ActionLabel:    "Instalar Vagrant",
		},
		{
			ID:             StepBox,
			Index:          4,
			Title:          "Añadir la caja del laboratorio",
			Why:            "Descargamos y registramos la 'caja' abxda/big-data-lab: una imagen Debian 11 con Hadoop 3.3.6, Kafka, Elasticsearch, Java y Jupyter ya instalados (≈4.4 GB). Esto no necesita administrador; Vagrant la guarda en tu carpeta de usuario.",
			NeedsElevation: false,
			ActionLabel:    "Añadir la caja",
		},
		{
			ID:             StepUp,
			Index:          5,
			Title:          "Levantar el entorno",
			Why:            "Generamos un Vagrantfile y ejecutamos 'vagrant up', que crea la máquina virtual en VirtualBox y la arranca. Seguimos el progreso en vivo. Al terminar tendrás una VM corriendo con todos los servicios listos para usar.",
			NeedsElevation: false,
			ActionLabel:    "Levantar VM (vagrant up)",
		},
		{
			ID:             StepServidores,
			Index:          6,
			Title:          "Iniciar servicios y ejercicio",
			Why:            "Este es el paso final. Abajo verás el estado de los servicios dentro de la VM (HDFS, Kafka, Elasticsearch, Jupyter) y el Ejercicio_01 (WordCount) para hacerlo paso a paso. Si los servicios están apagados, el botón 'Iniciar servicios' o el paso 1 del ejercicio los arrancan. Cuando todo esté en verde, usa los botones de arriba para abrir Jupyter o tu carpeta de trabajo. Es el mismo ejercicio que en la versión portable.",
			NeedsElevation: false,
			ActionLabel:    "Iniciar servicios",
		},
	}
}
