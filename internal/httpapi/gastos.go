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

// decorarGasto calcula el estado de pago que se sirve en la respuesta:
// "vencido" (pendiente + fecha de vencimiento pasada) se deriva al leer, no
// se guarda — igual que el estado de un contrato.
func decorarGasto(g *domain.Gasto) {
	g.EstadoPago = g.EstadoPagoDerivado(time.Now())
}

func validateGastoCampos(g domain.Gasto) string {
	if !g.Tipo.Valida() {
		return "tipo de gasto no válido: " + string(g.Tipo)
	}
	if g.Periodicidad != "" && !g.Periodicidad.Valida() {
		return "periodicidad no válida: " + string(g.Periodicidad)
	}
	if g.Importe <= 0 {
		return "el importe debe ser mayor que 0"
	}
	if time.Time(g.FechaEmision).IsZero() {
		return "el campo fechaEmision es obligatorio"
	}
	if g.EstadoPago != "" && g.EstadoPago != domain.PagoPendiente && g.EstadoPago != domain.PagoPagado && g.EstadoPago != domain.PagoVencido {
		return "estado de pago no válido: " + string(g.EstadoPago)
	}
	return ""
}

func handleListGastosInmueble(inmuebles *sqlite.InmueblesRepo, gastos *sqlite.GastosRepo) http.HandlerFunc {
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

		lista, err := gastos.ListByInmueble(inmuebleID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudieron listar los gastos")
			return
		}
		for i := range lista {
			decorarGasto(&lista[i])
		}
		writeJSON(w, http.StatusOK, lista)
	}
}

func handleCreateGastoInmueble(inmuebles *sqlite.InmueblesRepo, gastos *sqlite.GastosRepo) http.HandlerFunc {
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

		var g domain.Gasto
		if err := json.NewDecoder(r.Body).Decode(&g); err != nil {
			writeError(w, http.StatusBadRequest, "cuerpo de la petición inválido")
			return
		}
		if msg := validateGastoCampos(g); msg != "" {
			writeError(w, http.StatusBadRequest, msg)
			return
		}
		g.ID = 0
		g.InmuebleID = inmuebleID

		creado, err := gastos.Create(g)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudo crear el gasto")
			return
		}
		decorarGasto(&creado)
		writeJSON(w, http.StatusCreated, creado)
	}
}

func handleGetGasto(gastos *sqlite.GastosRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "id de gasto inválido")
			return
		}
		g, err := gastos.Get(id)
		if errors.Is(err, sqlite.ErrNotFound) {
			writeError(w, http.StatusNotFound, "gasto no encontrado")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudo leer el gasto")
			return
		}
		decorarGasto(&g)
		writeJSON(w, http.StatusOK, g)
	}
}

func handleUpdateGasto(gastos *sqlite.GastosRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "id de gasto inválido")
			return
		}
		actual, err := gastos.Get(id)
		if errors.Is(err, sqlite.ErrNotFound) {
			writeError(w, http.StatusNotFound, "gasto no encontrado")
			return
		} else if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudo leer el gasto")
			return
		}

		var g domain.Gasto
		if err := json.NewDecoder(r.Body).Decode(&g); err != nil {
			writeError(w, http.StatusBadRequest, "cuerpo de la petición inválido")
			return
		}
		if msg := validateGastoCampos(g); msg != "" {
			writeError(w, http.StatusBadRequest, msg)
			return
		}
		g.ID = id
		g.InmuebleID = actual.InmuebleID

		actualizado, err := gastos.Update(id, g)
		if errors.Is(err, sqlite.ErrNotFound) {
			writeError(w, http.StatusNotFound, "gasto no encontrado")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudo actualizar el gasto")
			return
		}
		decorarGasto(&actualizado)
		writeJSON(w, http.StatusOK, actualizado)
	}
}

// handleGetReciboGasto sirve el desglose por inquilino de una factura: busca
// el reparto vigente en la fecha de emisión del gasto para ese tipo de gasto
// y reparte el importe cuadrando el redondeo con el total (domain.CalcularRecibo).
// Un gasto sin reparto vigente (inmueble no compartido, o sin reparto para
// ese tipo) devuelve un recibo con sinReparto = true y sin líneas — 200, no
// error.
func handleGetReciboGasto(gastos *sqlite.GastosRepo, repartos *sqlite.RepartosRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "id de gasto inválido")
			return
		}
		g, err := gastos.Get(id)
		if errors.Is(err, sqlite.ErrNotFound) {
			writeError(w, http.StatusNotFound, "gasto no encontrado")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudo leer el gasto")
			return
		}

		cuotas, err := repartos.VigenteEnFecha(g.InmuebleID, g.Tipo, time.Time(g.FechaEmision))
		if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudo calcular el reparto")
			return
		}
		writeJSON(w, http.StatusOK, domain.CalcularRecibo(g, cuotas))
	}
}

func handleUploadDocumentoGasto(gastos *sqlite.GastosRepo, documentos *sqlite.DocumentosRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "id de gasto inválido")
			return
		}
		if _, err := gastos.Get(id); errors.Is(err, sqlite.ErrNotFound) {
			writeError(w, http.StatusNotFound, "gasto no encontrado")
			return
		} else if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudo comprobar el gasto")
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
			EntidadTipo:   domain.EntidadGasto,
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

func handleListDocumentosGasto(gastos *sqlite.GastosRepo, documentos *sqlite.DocumentosRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "id de gasto inválido")
			return
		}
		if _, err := gastos.Get(id); errors.Is(err, sqlite.ErrNotFound) {
			writeError(w, http.StatusNotFound, "gasto no encontrado")
			return
		} else if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudo comprobar el gasto")
			return
		}

		docs, err := documentos.ListByEntidad(domain.EntidadGasto, id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudieron listar los documentos")
			return
		}
		writeJSON(w, http.StatusOK, docs)
	}
}

// periodoMes traduce un "AAAA-MM" (o vacío = mes actual) al intervalo
// [primer día del mes, primer día del mes siguiente).
func periodoMes(q string) (etiqueta string, desde, hasta time.Time, err error) {
	var y int
	var m time.Month
	if q == "" {
		now := time.Now()
		y, m = now.Year(), now.Month()
	} else {
		t, e := time.Parse("2006-01", q)
		if e != nil {
			return "", time.Time{}, time.Time{}, fmt.Errorf("periodo %q inválido, se espera AAAA-MM", q)
		}
		y, m = t.Year(), t.Month()
	}
	desde = time.Date(y, m, 1, 0, 0, 0, 0, time.UTC)
	hasta = desde.AddDate(0, 1, 0)
	return desde.Format("2006-01"), desde, hasta, nil
}
