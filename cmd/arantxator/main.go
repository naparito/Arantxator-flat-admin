// Arantxator Flat Admin arranca un único proceso que sirve la GUI web,
// resuelve la API y abre la base de datos SQLite embebida — sin más
// dependencias externas que instalar.
package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"runtime"

	"github.com/naparito/Arantxator-flat-admin/internal/config"
	"github.com/naparito/Arantxator-flat-admin/internal/httpapi"
	"github.com/naparito/Arantxator-flat-admin/internal/storage/sqlite"
	"github.com/naparito/Arantxator-flat-admin/internal/webui"
)

func main() {
	cfg := config.Load()

	db, err := sqlite.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("no se pudo abrir la base de datos (%s): %v", cfg.DBPath, err)
	}
	defer db.Close()

	if err := sqlite.Migrate(db); err != nil {
		log.Fatalf("no se pudieron aplicar las migraciones: %v", err)
	}

	mux := http.NewServeMux()
	httpapi.RegisterRoutes(mux, db)
	mux.Handle("/", webui.Handler())

	url := "http://" + cfg.Addr
	// fmt.Println, no log.Printf: es un mensaje informativo, no un error.
	// log escribe por stderr por defecto, y algunas terminales (PowerShell
	// 5.1 con un proceso nativo) pintan cualquier línea de stderr como si
	// fuera un error, aunque el arranque haya ido bien.
	fmt.Printf("Arantxator Flat Admin escuchando en %s (base de datos: %s)\n", url, cfg.DBPath)
	go openBrowser(url)

	if err := http.ListenAndServe(cfg.Addr, mux); err != nil {
		log.Fatalf("no se pudo arrancar el servidor: %v", err)
	}
}

// openBrowser abre la pantalla principal en el navegador por defecto para
// que el usuario nunca vea una terminal ni tenga que teclear una URL.
func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}
