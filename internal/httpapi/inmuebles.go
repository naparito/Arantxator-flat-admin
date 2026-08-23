package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/naparito/Arantxator-flat-admin/internal/domain"
	"github.com/naparito/Arantxator-flat-admin/internal/storage/sqlite"
)

var tiposInmuebleValidos = map[domain.TipoInmueble]bool{
	domain.TipoPiso:       true,
	domain.TipoCasa:       true,
	domain.TipoHabitacion: true,
	domain.TipoLocal:      true,
}

var estadosInmuebleValidos = map[domain.EstadoInmueble]bool{
	domain.InmuebleDisponible:      true,
	domain.InmuebleAlquilado:       true,
	domain.InmuebleEnReforma:       true,
	domain.InmuebleFueraDeServicio: true,
}

func validateInmueble(m domain.Inmueble) string {
	if m.Nombre == "" {
		return "el campo nombre es obligatorio"
	}
	if m.Direccion == "" {
		return "el campo direccion es obligatorio"
	}
	if !tiposInmuebleValidos[m.Tipo] {
		return "tipo de inmueble no válido: " + string(m.Tipo)
	}
	if m.Estado != "" && !estadosInmuebleValidos[m.Estado] {
		return "estado de inmueble no válido: " + string(m.Estado)
	}
	return ""
}

func handleListInmuebles(repo *sqlite.InmueblesRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		estado := domain.EstadoInmueble(r.URL.Query().Get("estado"))
		if estado != "" && !estadosInmuebleValidos[estado] {
			writeError(w, http.StatusBadRequest, "estado de inmueble no válido: "+string(estado))
			return
		}
		inmuebles, err := repo.List(estado)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudieron listar los inmuebles")
			return
		}
		writeJSON(w, http.StatusOK, inmuebles)
	}
}

func handleCreateInmueble(repo *sqlite.InmueblesRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var m domain.Inmueble
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			writeError(w, http.StatusBadRequest, "cuerpo de la petición inválido")
			return
		}
		if msg := validateInmueble(m); msg != "" {
			writeError(w, http.StatusBadRequest, msg)
			return
		}
		creado, err := repo.Create(m)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudo crear el inmueble")
			return
		}
		writeJSON(w, http.StatusCreated, creado)
	}
}

func handleGetInmueble(repo *sqlite.InmueblesRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "id de inmueble inválido")
			return
		}
		m, err := repo.Get(id)
		if errors.Is(err, sqlite.ErrNotFound) {
			writeError(w, http.StatusNotFound, "inmueble no encontrado")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudo leer el inmueble")
			return
		}
		writeJSON(w, http.StatusOK, m)
	}
}

func handleUpdateInmueble(repo *sqlite.InmueblesRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "id de inmueble inválido")
			return
		}
		var m domain.Inmueble
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			writeError(w, http.StatusBadRequest, "cuerpo de la petición inválido")
			return
		}
		if msg := validateInmueble(m); msg != "" {
			writeError(w, http.StatusBadRequest, msg)
			return
		}
		actualizado, err := repo.Update(id, m)
		if errors.Is(err, sqlite.ErrNotFound) {
			writeError(w, http.StatusNotFound, "inmueble no encontrado")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudo actualizar el inmueble")
			return
		}
		writeJSON(w, http.StatusOK, actualizado)
	}
}
