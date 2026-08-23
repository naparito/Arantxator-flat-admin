package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/naparito/Arantxator-flat-admin/internal/domain"
	"github.com/naparito/Arantxator-flat-admin/internal/storage/sqlite"
)

func validateInquilino(i domain.Inquilino) string {
	if i.NombreCompleto == "" {
		return "el campo nombreCompleto es obligatorio"
	}
	if i.DocumentoIdentidad == "" {
		return "el campo documentoIdentidad es obligatorio"
	}
	return ""
}

func handleListInquilinos(repo *sqlite.InquilinosRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		inquilinos, err := repo.List()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudieron listar los inquilinos")
			return
		}
		writeJSON(w, http.StatusOK, inquilinos)
	}
}

func handleCreateInquilino(repo *sqlite.InquilinosRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var i domain.Inquilino
		if err := json.NewDecoder(r.Body).Decode(&i); err != nil {
			writeError(w, http.StatusBadRequest, "cuerpo de la petición inválido")
			return
		}
		if msg := validateInquilino(i); msg != "" {
			writeError(w, http.StatusBadRequest, msg)
			return
		}
		creado, err := repo.Create(i)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudo crear el inquilino")
			return
		}
		writeJSON(w, http.StatusCreated, creado)
	}
}

func handleGetInquilino(repo *sqlite.InquilinosRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "id de inquilino inválido")
			return
		}
		i, err := repo.Get(id)
		if errors.Is(err, sqlite.ErrNotFound) {
			writeError(w, http.StatusNotFound, "inquilino no encontrado")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudo leer el inquilino")
			return
		}
		writeJSON(w, http.StatusOK, i)
	}
}

func handleUpdateInquilino(repo *sqlite.InquilinosRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "id de inquilino inválido")
			return
		}
		var i domain.Inquilino
		if err := json.NewDecoder(r.Body).Decode(&i); err != nil {
			writeError(w, http.StatusBadRequest, "cuerpo de la petición inválido")
			return
		}
		if msg := validateInquilino(i); msg != "" {
			writeError(w, http.StatusBadRequest, msg)
			return
		}
		actualizado, err := repo.Update(id, i)
		if errors.Is(err, sqlite.ErrNotFound) {
			writeError(w, http.StatusNotFound, "inquilino no encontrado")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudo actualizar el inquilino")
			return
		}
		writeJSON(w, http.StatusOK, actualizado)
	}
}
