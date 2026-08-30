package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/naparito/Arantxator-flat-admin/internal/domain"
	"github.com/naparito/Arantxator-flat-admin/internal/storage/sqlite"
)

func validateHabitacion(h domain.Habitacion) string {
	if h.Nombre == "" {
		return "el campo nombre es obligatorio"
	}
	return ""
}

func handleListHabitaciones(inmuebles *sqlite.InmueblesRepo, habitaciones *sqlite.HabitacionesRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		inmuebleID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "id de inmueble inválido")
			return
		}
		if _, err := inmuebles.Get(inmuebleID); errors.Is(err, sqlite.ErrNotFound) {
			writeError(w, http.StatusNotFound, "inmueble no encontrado")
			return
		} else if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudo comprobar el inmueble")
			return
		}

		lista, err := habitaciones.ListByInmueble(inmuebleID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudieron listar las habitaciones")
			return
		}
		writeJSON(w, http.StatusOK, lista)
	}
}

func handleCreateHabitacion(inmuebles *sqlite.InmueblesRepo, habitaciones *sqlite.HabitacionesRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		inmuebleID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "id de inmueble inválido")
			return
		}
		if _, err := inmuebles.Get(inmuebleID); errors.Is(err, sqlite.ErrNotFound) {
			writeError(w, http.StatusNotFound, "inmueble no encontrado")
			return
		} else if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudo comprobar el inmueble")
			return
		}

		var h domain.Habitacion
		if err := json.NewDecoder(r.Body).Decode(&h); err != nil {
			writeError(w, http.StatusBadRequest, "cuerpo de la petición inválido")
			return
		}
		if msg := validateHabitacion(h); msg != "" {
			writeError(w, http.StatusBadRequest, msg)
			return
		}
		h.InmuebleID = inmuebleID

		creada, err := habitaciones.Create(h)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudo crear la habitación")
			return
		}
		writeJSON(w, http.StatusCreated, creada)
	}
}

func handleGetHabitacion(habitaciones *sqlite.HabitacionesRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "id de habitación inválido")
			return
		}
		h, err := habitaciones.Get(id)
		if errors.Is(err, sqlite.ErrNotFound) {
			writeError(w, http.StatusNotFound, "habitación no encontrada")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudo leer la habitación")
			return
		}
		writeJSON(w, http.StatusOK, h)
	}
}

func handleUpdateHabitacion(habitaciones *sqlite.HabitacionesRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "id de habitación inválido")
			return
		}
		var h domain.Habitacion
		if err := json.NewDecoder(r.Body).Decode(&h); err != nil {
			writeError(w, http.StatusBadRequest, "cuerpo de la petición inválido")
			return
		}
		if msg := validateHabitacion(h); msg != "" {
			writeError(w, http.StatusBadRequest, msg)
			return
		}
		actualizada, err := habitaciones.Update(id, h)
		if errors.Is(err, sqlite.ErrNotFound) {
			writeError(w, http.StatusNotFound, "habitación no encontrada")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudo actualizar la habitación")
			return
		}
		writeJSON(w, http.StatusOK, actualizada)
	}
}

type asignarOcupanteRequest struct {
	InquilinoID *int64 `json:"inquilinoId"`
}

// handleAsignarOcupanteHabitacion asigna o quita (inquilinoId nulo) el
// inquilino ocupante de una habitación. Es una asignación de presentación
// (quién vive hoy ahí), independiente del contrato — ver domain.Habitacion.
func handleAsignarOcupanteHabitacion(habitaciones *sqlite.HabitacionesRepo, inquilinos *sqlite.InquilinosRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "id de habitación inválido")
			return
		}
		if _, err := habitaciones.Get(id); errors.Is(err, sqlite.ErrNotFound) {
			writeError(w, http.StatusNotFound, "habitación no encontrada")
			return
		} else if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudo comprobar la habitación")
			return
		}

		var body asignarOcupanteRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "cuerpo de la petición inválido")
			return
		}

		if body.InquilinoID != nil {
			if _, err := inquilinos.Get(*body.InquilinoID); errors.Is(err, sqlite.ErrNotFound) {
				writeError(w, http.StatusNotFound, "inquilino no encontrado")
				return
			} else if err != nil {
				writeError(w, http.StatusInternalServerError, "no se pudo comprobar el inquilino")
				return
			}
		}

		actualizada, err := habitaciones.AsignarOcupante(id, body.InquilinoID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudo asignar el ocupante")
			return
		}
		writeJSON(w, http.StatusOK, actualizada)
	}
}

func handleDeleteHabitacion(habitaciones *sqlite.HabitacionesRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "id de habitación inválido")
			return
		}
		err = habitaciones.Delete(id)
		if errors.Is(err, sqlite.ErrNotFound) {
			writeError(w, http.StatusNotFound, "habitación no encontrada")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudo borrar la habitación")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
