package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/naparito/Arantxator-flat-admin/internal/domain"
	"github.com/naparito/Arantxator-flat-admin/internal/storage/sqlite"
)

var estadosFianzaValidos = map[domain.EstadoFianza]bool{
	domain.FianzaPendiente:    true,
	domain.FianzaDepositada:   true,
	domain.FianzaEnDevolucion: true,
	domain.FianzaDevuelta:     true,
}

var estadosContratoValidos = map[domain.EstadoContrato]bool{
	domain.ContratoActivo:         true,
	domain.ContratoProximoAVencer: true,
	domain.ContratoVencido:        true,
	domain.ContratoRescindido:     true,
}

// decorarContrato calcula los campos derivados que se sirven en la respuesta
// pero no se guardan: el estado (activo / próximo a vencer / vencido) a
// partir de las fechas, y la fecha límite de depósito de la fianza.
func decorarContrato(c *domain.Contrato) {
	c.Estado = c.EstadoDerivado(time.Now())
	c.FechaLimiteDepositoFianza = domain.Fecha(domain.FechaLimiteDepositoFianza(time.Time(c.FechaFirma)))
}

func validateContratoCampos(c domain.Contrato) string {
	if c.InmuebleID <= 0 {
		return "el campo inmuebleId es obligatorio"
	}
	if len(c.InquilinoIDs) == 0 {
		return "un contrato necesita al menos un inquilino (co-arrendatario)"
	}
	if time.Time(c.FechaFirma).IsZero() {
		return "el campo fechaFirma es obligatorio"
	}
	if time.Time(c.FechaInicio).IsZero() {
		return "el campo fechaInicio es obligatorio"
	}
	if time.Time(c.FechaFin).IsZero() {
		return "el campo fechaFin es obligatorio"
	}
	if !time.Time(c.FechaFin).After(time.Time(c.FechaInicio)) {
		return "la fecha de fin debe ser posterior a la de inicio"
	}
	if c.RentaMensual <= 0 {
		return "la renta mensual debe ser mayor que 0"
	}
	if c.FianzaEstado != "" && !estadosFianzaValidos[c.FianzaEstado] {
		return "estado de fianza no válido: " + string(c.FianzaEstado)
	}
	if c.Estado != "" && !estadosContratoValidos[c.Estado] {
		return "estado de contrato no válido: " + string(c.Estado)
	}
	return ""
}

// validateContratoAmbito comprueba las reglas que dependen del inmueble: en un
// inmueble compartido el contrato es de una habitación concreta (habitacionId
// obligatorio y de ese inmueble); en uno no compartido, del inmueble entero
// (habitacionId debe ser nulo).
func validateContratoAmbito(c domain.Contrato, inmueble domain.Inmueble, habitaciones *sqlite.HabitacionesRepo) (int, string) {
	if inmueble.Compartido {
		if c.HabitacionID == nil {
			return http.StatusBadRequest, "el inmueble es compartido: hay que elegir la habitación del contrato"
		}
		hab, err := habitaciones.Get(*c.HabitacionID)
		if errors.Is(err, sqlite.ErrNotFound) {
			return http.StatusBadRequest, fmt.Sprintf("la habitación %d no existe", *c.HabitacionID)
		}
		if err != nil {
			return http.StatusInternalServerError, "no se pudo comprobar la habitación"
		}
		if hab.InmuebleID != inmueble.ID {
			return http.StatusBadRequest, "la habitación no pertenece a ese inmueble"
		}
		return 0, ""
	}
	if c.HabitacionID != nil {
		return http.StatusBadRequest, "el inmueble no es compartido: el contrato no puede ir asociado a una habitación"
	}
	return 0, ""
}

func comprobarInquilinos(ids []int64, inquilinos *sqlite.InquilinosRepo) (int, string) {
	for _, id := range ids {
		if _, err := inquilinos.Get(id); errors.Is(err, sqlite.ErrNotFound) {
			return http.StatusBadRequest, fmt.Sprintf("el inquilino %d no existe", id)
		} else if err != nil {
			return http.StatusInternalServerError, "no se pudo comprobar el inquilino"
		}
	}
	return 0, ""
}

func handleListContratos(contratos *sqlite.ContratosRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lista, err := contratos.List()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudieron listar los contratos")
			return
		}
		for i := range lista {
			decorarContrato(&lista[i])
		}
		writeJSON(w, http.StatusOK, lista)
	}
}

func handleGetContrato(contratos *sqlite.ContratosRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "id de contrato inválido")
			return
		}
		c, err := contratos.Get(id)
		if errors.Is(err, sqlite.ErrNotFound) {
			writeError(w, http.StatusNotFound, "contrato no encontrado")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudo leer el contrato")
			return
		}
		decorarContrato(&c)
		writeJSON(w, http.StatusOK, c)
	}
}

func handleCreateContrato(contratos *sqlite.ContratosRepo, inmuebles *sqlite.InmueblesRepo, habitaciones *sqlite.HabitacionesRepo, inquilinos *sqlite.InquilinosRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var c domain.Contrato
		if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
			writeError(w, http.StatusBadRequest, "cuerpo de la petición inválido")
			return
		}
		c.ID = 0
		if msg := validateContratoCampos(c); msg != "" {
			writeError(w, http.StatusBadRequest, msg)
			return
		}

		inmueble, err := inmuebles.Get(c.InmuebleID)
		if errors.Is(err, sqlite.ErrNotFound) {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("el inmueble %d no existe", c.InmuebleID))
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudo comprobar el inmueble")
			return
		}

		if code, msg := validateContratoAmbito(c, inmueble, habitaciones); msg != "" {
			writeError(w, code, msg)
			return
		}
		if code, msg := comprobarInquilinos(c.InquilinoIDs, inquilinos); msg != "" {
			writeError(w, code, msg)
			return
		}

		solapa, err := contratos.TieneContratoVigenteEnAmbito(c.InmuebleID, c.HabitacionID, 0, time.Now())
		if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudo comprobar el solapamiento de contratos")
			return
		}
		if solapa {
			writeError(w, http.StatusConflict, ambitoOcupadoMsg(c.HabitacionID))
			return
		}

		creado, err := contratos.Create(c)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudo crear el contrato")
			return
		}
		decorarContrato(&creado)
		writeJSON(w, http.StatusCreated, creado)
	}
}

func handleUpdateContrato(contratos *sqlite.ContratosRepo, inmuebles *sqlite.InmueblesRepo, habitaciones *sqlite.HabitacionesRepo, inquilinos *sqlite.InquilinosRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "id de contrato inválido")
			return
		}
		if _, err := contratos.Get(id); errors.Is(err, sqlite.ErrNotFound) {
			writeError(w, http.StatusNotFound, "contrato no encontrado")
			return
		} else if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudo leer el contrato")
			return
		}

		var c domain.Contrato
		if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
			writeError(w, http.StatusBadRequest, "cuerpo de la petición inválido")
			return
		}
		c.ID = id
		if msg := validateContratoCampos(c); msg != "" {
			writeError(w, http.StatusBadRequest, msg)
			return
		}

		inmueble, err := inmuebles.Get(c.InmuebleID)
		if errors.Is(err, sqlite.ErrNotFound) {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("el inmueble %d no existe", c.InmuebleID))
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudo comprobar el inmueble")
			return
		}
		if code, msg := validateContratoAmbito(c, inmueble, habitaciones); msg != "" {
			writeError(w, code, msg)
			return
		}
		if code, msg := comprobarInquilinos(c.InquilinoIDs, inquilinos); msg != "" {
			writeError(w, code, msg)
			return
		}

		// Un contrato rescindido ya no ocupa su ámbito, así que no cuenta
		// para el solapamiento.
		if c.Estado != domain.ContratoRescindido {
			solapa, err := contratos.TieneContratoVigenteEnAmbito(c.InmuebleID, c.HabitacionID, id, time.Now())
			if err != nil {
				writeError(w, http.StatusInternalServerError, "no se pudo comprobar el solapamiento de contratos")
				return
			}
			if solapa {
				writeError(w, http.StatusConflict, ambitoOcupadoMsg(c.HabitacionID))
				return
			}
		}

		actualizado, err := contratos.Update(id, c)
		if errors.Is(err, sqlite.ErrNotFound) {
			writeError(w, http.StatusNotFound, "contrato no encontrado")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudo actualizar el contrato")
			return
		}
		decorarContrato(&actualizado)
		writeJSON(w, http.StatusOK, actualizado)
	}
}

func ambitoOcupadoMsg(habitacionID *int64) string {
	if habitacionID != nil {
		return "esa habitación ya tiene un contrato vigente"
	}
	return "ese inmueble ya tiene un contrato vigente"
}

func handleListContratosInquilino(contratos *sqlite.ContratosRepo, inquilinos *sqlite.InquilinosRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "id de inquilino inválido")
			return
		}
		if _, err := inquilinos.Get(id); errors.Is(err, sqlite.ErrNotFound) {
			writeError(w, http.StatusNotFound, "inquilino no encontrado")
			return
		} else if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudo comprobar el inquilino")
			return
		}

		lista, err := contratos.ListByInquilino(id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudieron listar los contratos del inquilino")
			return
		}
		for i := range lista {
			decorarContrato(&lista[i])
		}
		writeJSON(w, http.StatusOK, lista)
	}
}

func handleUploadDocumentoContrato(contratos *sqlite.ContratosRepo, documentos *sqlite.DocumentosRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "id de contrato inválido")
			return
		}
		if _, err := contratos.Get(id); errors.Is(err, sqlite.ErrNotFound) {
			writeError(w, http.StatusNotFound, "contrato no encontrado")
			return
		} else if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudo comprobar el contrato")
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
			EntidadTipo:   domain.EntidadContrato,
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

func handleListDocumentosContrato(contratos *sqlite.ContratosRepo, documentos *sqlite.DocumentosRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "id de contrato inválido")
			return
		}
		if _, err := contratos.Get(id); errors.Is(err, sqlite.ErrNotFound) {
			writeError(w, http.StatusNotFound, "contrato no encontrado")
			return
		} else if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudo comprobar el contrato")
			return
		}

		docs, err := documentos.ListByEntidad(domain.EntidadContrato, id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudieron listar los documentos")
			return
		}
		writeJSON(w, http.StatusOK, docs)
	}
}
