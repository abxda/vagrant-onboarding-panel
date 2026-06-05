//go:build linux

package main

import "os"

// vboxdrvLoaded indica si el módulo de kernel host de VirtualBox (vboxdrv) está
// cargado. `vagrant up --provider virtualbox` falla si NO lo está (típico con
// Secure Boot, que bloquea el módulo sin firmar). Comprobamos /sys/module/vboxdrv,
// que existe sólo cuando el módulo está cargado en el kernel.
func vboxdrvLoaded() bool {
	_, err := os.Stat("/sys/module/vboxdrv")
	return err == nil
}
