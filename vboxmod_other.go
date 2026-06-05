//go:build !linux

package main

// vboxdrvLoaded: en Windows y macOS, VirtualBox NO usa un módulo de kernel
// estilo Linux (vboxdrv), así que no hay nada que comprobar ni cargar. Devolvemos
// true para que el flujo de arranque continúe igual que antes en esas plataformas.
func vboxdrvLoaded() bool { return true }
