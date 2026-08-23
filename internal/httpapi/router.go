// Package httpapi expone la API HTTP interna que consume la SPA. Empieza
// con un único endpoint de salud; los módulos de Inmuebles, Inquilinos,
// Contratos, Gastos e Incidencias se añaden en iteraciones posteriores.
package httpapi

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux, db *sql.DB) {
	mux.HandleFunc("/api/health", handleHealth(db))
}

func handleHealth(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := "ok"
		if err := db.Ping(); err != nil {
			status = "db_error"
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": status})
	}
}
