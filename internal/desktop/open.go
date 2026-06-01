// Package desktop abre rutas y URLs en las apps nativas del host (explorador
// de archivos / navegador), de forma multiplataforma. Se usa para que el
// alumno abra su carpeta de trabajo y Jupyter desde el panel.
package desktop

import (
	"os/exec"
	"runtime"
)

// OpenPath abre una carpeta o archivo en el explorador del sistema.
//
//	Windows -> explorer
//	macOS   -> open
//	Linux   -> xdg-open
//
// explorer.exe suele devolver un código de salida distinto de cero aunque
// tenga éxito; como lanzamos con Start() (sin esperar) eso no nos afecta.
func OpenPath(path string) error {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("explorer", path).Start()
	case "darwin":
		return exec.Command("open", path).Start()
	default:
		return exec.Command("xdg-open", path).Start()
	}
}

// OpenURL abre una URL en el navegador por defecto del sistema.
func OpenURL(url string) error {
	switch runtime.GOOS {
	case "windows":
		// rundll32 evita problemas de parsing de & en URLs que tiene `start`.
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	default:
		return exec.Command("xdg-open", url).Start()
	}
}
