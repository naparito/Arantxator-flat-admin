package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/naparito/Arantxator-flat-admin/internal/domain"
	"github.com/naparito/Arantxator-flat-admin/internal/storage/sqlite"
)

// validateIncidenciaCampos comprueba los campos que no dependen del estado
// actual (los comunes a alta y edición).
func validateIncidenciaCampos(inc domain.Incidencia) string {
	if inc.Titulo == "" {
		return "el campo titulo es obligatorio"
	}
	if inc.Categoria == "" {
		return "el campo categoria es obligatorio"
	}
	if inc.Prioridad != "" && !inc.Prioridad.Valida() {
		return "prioridad no válida: " + string(inc.Prioridad)
	}
	if inc.Origen != "" && !inc.Origen.Valido() {
		return "origen no válido: " + string(inc.Origen)
	}
	if inc.CosteACargoDe != "" && !inc.CosteACargoDe.Valido() {
		return "coste a cargo de no válido: " + string(inc.CosteACargoDe)
	}
	if inc.Coste < 0 {
		return "el coste no puede ser negativo"
	}
	return ""
}

func handleListIncidenciasInmueble(inmuebles *sqlite.InmueblesRepo, incidencias *sqlite.IncidenciasRepo) http.HandlerFunc {
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

		lista, err := incidencias.ListByInmueble(inmuebleID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudieron listar las incidencias")
			return
		}
		writeJSON(w, http.StatusOK, lista)
	}
}

func handleCreateIncidenciaInmueble(inmuebles *sqlite.InmueblesRepo, incidencias *sqlite.IncidenciasRepo) http.HandlerFunc {
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

		var inc domain.Incidencia
		if err := json.NewDecoder(r.Body).Decode(&inc); err != nil {
			writeError(w, http.StatusBadRequest, "cuerpo de la petición inválido")
			return
		}
		if msg := validateIncidenciaCampos(inc); msg != "" {
			writeError(w, http.StatusBadRequest, msg)
			return
		}
		inc.ID = 0
		inc.InmuebleID = inmuebleID
		// El alta siempre entra "abierta"; el flujo se recorre después con PUT.
		inc.Estado = domain.IncidenciaAbierta

		creada, err := incidencias.Create(inc)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudo crear la incidencia")
			return
		}
		writeJSON(w, http.StatusCreated, creada)
	}
}

func handleGetIncidencia(incidencias *sqlite.IncidenciasRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "id de incidencia inválido")
			return
		}
		inc, err := incidencias.Get(id)
		if errors.Is(err, sqlite.ErrNotFound) {
			writeError(w, http.StatusNotFound, "incidencia no encontrada")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudo leer la incidencia")
			return
		}
		writeJSON(w, http.StatusOK, inc)
	}
}

func handleUpdateIncidencia(incidencias *sqlite.IncidenciasRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "id de incidencia inválido")
			return
		}
		actual, err := incidencias.Get(id)
		if errors.Is(err, sqlite.ErrNotFound) {
			writeError(w, http.StatusNotFound, "incidencia no encontrada")
			return
		} else if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudo leer la incidencia")
			return
		}

		var inc domain.Incidencia
		if err := json.NewDecoder(r.Body).Decode(&inc); err != nil {
			writeError(w, http.StatusBadRequest, "cuerpo de la petición inválido")
			return
		}
		if msg := validateIncidenciaCampos(inc); msg != "" {
			writeError(w, http.StatusBadRequest, msg)
			return
		}
		inc.ID = id
		inc.InmuebleID = actual.InmuebleID

		// El estado se mueve a mano por el flujo abierta → en proceso →
		// esperando proveedor → resuelta → cerrada. Un estado vacío en el
		// cuerpo significa "no lo toques".
		if inc.Estado == "" {
			inc.Estado = actual.Estado
		}
		if inc.Estado != actual.Estado {
			if !inc.Estado.Valido() {
				writeError(w, http.StatusBadRequest, "estado de incidencia no válido: "+string(inc.Estado))
				return
			}
			if !domain.TransicionEstadoIncidenciaValida(actual.Estado, inc.Estado) {
				writeError(w, http.StatusConflict,
					"transición de estado no permitida: "+string(actual.Estado)+" → "+string(inc.Estado))
				return
			}
		}

		actualizada, err := incidencias.Update(id, inc)
		if errors.Is(err, sqlite.ErrNotFound) {
			writeError(w, http.StatusNotFound, "incidencia no encontrada")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudo actualizar la incidencia")
			return
		}
		writeJSON(w, http.StatusOK, actualizada)
	}
}

func handleUploadDocumentoIncidencia(incidencias *sqlite.IncidenciasRepo, documentos *sqlite.DocumentosRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "id de incidencia inválido")
			return
		}
		if _, err := incidencias.Get(id); errors.Is(err, sqlite.ErrNotFound) {
			writeError(w, http.StatusNotFound, "incidencia no encontrada")
			return
		} else if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudo comprobar la incidencia")
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxDocumentoBytes)
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			writeError(w, http.StatusBadRequest, "no se pudo leer el documento subido (¿supera el tamaño máximo?)")
			return
		}
		file, header, err := r.FormFile("archivo")
		if err != nil {
			writeError(w, http.StatusBadRequest, "falta el campo 'archivo' con el fichero a subir")
			return
		}
		defer file.Close()

		contenido, err := io.ReadAll(file)
		if err != nil {
			writeError(w, http.StatusBadRequest, "no se pudo leer el contenido del fichero")
			return
		}

		tipoMime := header.Header.Get("Content-Type")
		if tipoMime == "" {
			tipoMime = "application/octet-stream"
		}

		doc, err := documentos.Create(domain.Documento{
			EntidadTipo:   domain.EntidadIncidencia,
			EntidadID:     id,
			NombreArchivo: header.Filename,
			TipoMIME:      tipoMime,
			Contenido:     contenido,
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudo guardar el documento")
			return
		}
		writeJSON(w, http.StatusCreated, doc)
	}
}

func handleListDocumentosIncidencia(incidencias *sqlite.IncidenciasRepo, documentos *sqlite.DocumentosRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "id de incidencia inválido")
			return
		}
		if _, err := incidencias.Get(id); errors.Is(err, sqlite.ErrNotFound) {
			writeError(w, http.StatusNotFound, "incidencia no encontrada")
			return
		} else if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudo comprobar la incidencia")
			return
		}

		docs, err := documentos.ListByEntidad(domain.EntidadIncidencia, id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudieron listar los documentos")
			return
		}
		writeJSON(w, http.StatusOK, docs)
	}
}
