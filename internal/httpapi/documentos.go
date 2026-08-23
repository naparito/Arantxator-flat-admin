package httpapi

import (
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/naparito/Arantxator-flat-admin/internal/domain"
	"github.com/naparito/Arantxator-flat-admin/internal/storage/sqlite"
)

// maxDocumentoBytes limita el tamaño de un documento subido (escrituras,
// fotos en alta resolución…) para no agotar la memoria del proceso.
const maxDocumentoBytes = 50 << 20 // 50 MiB

func handleUploadDocumentoInmueble(inmuebles *sqlite.InmueblesRepo, documentos *sqlite.DocumentosRepo) http.HandlerFunc {
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
			EntidadTipo:   domain.EntidadInmueble,
			EntidadID:     inmuebleID,
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

func handleListDocumentosInmueble(inmuebles *sqlite.InmueblesRepo, documentos *sqlite.DocumentosRepo) http.HandlerFunc {
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

		docs, err := documentos.ListByEntidad(domain.EntidadInmueble, inmuebleID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudieron listar los documentos")
			return
		}
		writeJSON(w, http.StatusOK, docs)
	}
}

func handleUploadDocumentoInquilino(inquilinos *sqlite.InquilinosRepo, documentos *sqlite.DocumentosRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		inquilinoID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "id de inquilino inválido")
			return
		}
		if _, err := inquilinos.Get(inquilinoID); errors.Is(err, sqlite.ErrNotFound) {
			writeError(w, http.StatusNotFound, "inquilino no encontrado")
			return
		} else if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudo comprobar el inquilino")
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
			EntidadTipo:   domain.EntidadInquilino,
			EntidadID:     inquilinoID,
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

func handleListDocumentosInquilino(inquilinos *sqlite.InquilinosRepo, documentos *sqlite.DocumentosRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		inquilinoID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "id de inquilino inválido")
			return
		}
		if _, err := inquilinos.Get(inquilinoID); errors.Is(err, sqlite.ErrNotFound) {
			writeError(w, http.StatusNotFound, "inquilino no encontrado")
			return
		} else if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudo comprobar el inquilino")
			return
		}

		docs, err := documentos.ListByEntidad(domain.EntidadInquilino, inquilinoID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudieron listar los documentos")
			return
		}
		writeJSON(w, http.StatusOK, docs)
	}
}

func handleGetDocumento(documentos *sqlite.DocumentosRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "id de documento inválido")
			return
		}
		doc, err := documentos.Get(id)
		if errors.Is(err, sqlite.ErrNotFound) {
			writeError(w, http.StatusNotFound, "documento no encontrado")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudo leer el documento")
			return
		}
		w.Header().Set("Content-Type", doc.TipoMIME)
		w.Header().Set("Content-Disposition", `inline; filename="`+doc.NombreArchivo+`"`)
		w.Write(doc.Contenido)
	}
}
