// Package webui embebe la GUI web compilada dentro del propio binario, de
// forma que un único ejecutable ya lleva la interfaz consigo. Hasta que
// exista la SPA, sirve la página mínima de ./dist.
package webui

import (
	"embed"
	"io"
	"io/fs"
	"net/http"
	"strings"
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

// Handler sirve los ficheros ya compilados de la SPA y, para cualquier ruta
// que no corresponda a un fichero real (las rutas de React Router como
// /inmuebles/5), cae a index.html — si no, recargar la página en una ruta
// interna (o abrir un enlace directo a una ficha) daría 404 en vez de dejar
// que el router de la SPA la resuelva en el cliente.
//
// No basta con delegar en http.FileServer apuntando a "/index.html": ese
// handler trata "index.html" como especial y hace un 301 a "/", perdiendo
// la ruta que el router de React necesita para renderizar la pantalla
// correcta. Por eso el fallback sirve el contenido directamente.
func Handler() http.Handler {
	fileServer := http.FileServer(http.FS(Assets))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}
		if _, err := fs.Stat(Assets, path); err != nil {
			serveIndex(w, r)
			return
		}
		fileServer.ServeHTTP(w, r)
	})
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	f, err := Assets.Open("index.html")
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		http.Error(w, "no se pudo servir la aplicación", http.StatusInternalServerError)
		return
	}
	rs, ok := f.(io.ReadSeeker)
	if !ok {
		http.Error(w, "no se pudo servir la aplicación", http.StatusInternalServerError)
		return
	}
	http.ServeContent(w, r, "index.html", stat.ModTime(), rs)
}
