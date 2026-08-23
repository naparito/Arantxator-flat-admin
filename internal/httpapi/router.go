// Package httpapi expone la API HTTP interna que consume la SPA. El módulo
// de Inmuebles es el primero; Inquilinos, Contratos, Gastos e Incidencias
// se añaden en iteraciones posteriores.
package httpapi

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/naparito/Arantxator-flat-admin/internal/storage/sqlite"
)

func RegisterRoutes(mux *http.ServeMux, db *sql.DB) {
	mux.HandleFunc("/api/health", handleHealth(db))

	inmuebles := sqlite.NewInmueblesRepo(db)
	documentos := sqlite.NewDocumentosRepo(db)
	habitaciones := sqlite.NewHabitacionesRepo(db)

	mux.HandleFunc("GET /api/inmuebles", handleListInmuebles(inmuebles))
	mux.HandleFunc("POST /api/inmuebles", handleCreateInmueble(inmuebles))
	mux.HandleFunc("GET /api/inmuebles/{id}", handleGetInmueble(inmuebles))
	mux.HandleFunc("PUT /api/inmuebles/{id}", handleUpdateInmueble(inmuebles))
	mux.HandleFunc("POST /api/inmuebles/{id}/documentos", handleUploadDocumentoInmueble(inmuebles, documentos))
	mux.HandleFunc("GET /api/inmuebles/{id}/documentos", handleListDocumentosInmueble(inmuebles, documentos))
	mux.HandleFunc("GET /api/documentos/{id}", handleGetDocumento(documentos))
	mux.HandleFunc("GET /api/inmuebles/{id}/habitaciones", handleListHabitaciones(inmuebles, habitaciones))
	mux.HandleFunc("POST /api/inmuebles/{id}/habitaciones", handleCreateHabitacion(inmuebles, habitaciones))
	mux.HandleFunc("GET /api/habitaciones/{id}", handleGetHabitacion(habitaciones))
	mux.HandleFunc("PUT /api/habitaciones/{id}", handleUpdateHabitacion(habitaciones))
	mux.HandleFunc("DELETE /api/habitaciones/{id}", handleDeleteHabitacion(habitaciones))
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
