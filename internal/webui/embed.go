// Package webui embebe la GUI web compilada dentro del propio binario, de
// forma que un único ejecutable ya lleva la interfaz consigo. Hasta que
// exista la SPA, sirve la página mínima de ./dist.
package webui

import (
	"embed"
	"io/fs"
)

//go:embed dist
var embedded embed.FS

// Assets es el sistema de ficheros de la GUI ya compilada, listo para
// servirse con http.FileServer.
var Assets fs.FS

func init() {
	sub, err := fs.Sub(embedded, "dist")
	if err != nil {
		panic(err)
	}
	Assets = sub
}
