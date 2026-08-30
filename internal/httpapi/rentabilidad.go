package httpapi

import (
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/naparito/Arantxator-flat-admin/internal/domain"
	"github.com/naparito/Arantxator-flat-admin/internal/storage/sqlite"
)

// handleGetRentabilidadInmueble sirve la rentabilidad neta de un inmueble en
// un mes: ingresos (renta cobrada, tabla cobros_renta) − gastos (facturas
// con fecha de emisión en el mes) = neto. El periodo llega como ?periodo=AAAA-MM
// (por defecto, el mes actual). Se calcula al leer sobre los datos reales;
// el frontend no cruza cobros y gastos por su cuenta.
func handleGetRentabilidadInmueble(inmuebles *sqlite.InmueblesRepo, gastos *sqlite.GastosRepo, cobros *sqlite.CobrosRepo) http.HandlerFunc {
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

		etiqueta, desde, hasta, err := periodoMes(r.URL.Query().Get("periodo"))
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		ingresos, err := cobros.SumaEnPeriodo(inmuebleID, desde, hasta)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudieron sumar los ingresos")
			return
		}
		gastoTotal, err := gastos.SumaImporteEnPeriodo(inmuebleID, desde, hasta)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudieron sumar los gastos")
			return
		}

		writeJSON(w, http.StatusOK, domain.Rentabilidad{
			InmuebleID: inmuebleID,
			Periodo:    etiqueta,
			Ingresos:   redondea2(ingresos),
			Gastos:     redondea2(gastoTotal),
			Neto:       redondea2(ingresos - gastoTotal),
		})
	}
}

func redondea2(x float64) float64 {
	return math.Round(x*100) / 100
}

func validateCobroCampos(c domain.CobroRenta) string {
	if c.Importe <= 0 {
		return "el importe del cobro debe ser mayor que 0"
	}
	if time.Time(c.Periodo).IsZero() {
		return "el campo periodo es obligatorio (AAAA-MM-DD, primer día del mes)"
	}
	return ""
}

func handleListCobrosInmueble(inmuebles *sqlite.InmueblesRepo, cobros *sqlite.CobrosRepo) http.HandlerFunc {
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
		lista, err := cobros.ListByInmueble(inmuebleID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudieron listar los cobros")
			return
		}
		writeJSON(w, http.StatusOK, lista)
	}
}

func handleCreateCobroInmueble(inmuebles *sqlite.InmueblesRepo, cobros *sqlite.CobrosRepo) http.HandlerFunc {
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

		var c domain.CobroRenta
		if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
			writeError(w, http.StatusBadRequest, "cuerpo de la petición inválido")
			return
		}
		if msg := validateCobroCampos(c); msg != "" {
			writeError(w, http.StatusBadRequest, msg)
			return
		}
		c.ID = 0
		c.InmuebleID = inmuebleID

		creado, err := cobros.Create(c)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudo registrar el cobro")
			return
		}
		writeJSON(w, http.StatusCreated, creado)
	}
}

func handleUpdateCobro(cobros *sqlite.CobrosRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "id de cobro inválido")
			return
		}
		if _, err := cobros.Get(id); errors.Is(err, sqlite.ErrNotFound) {
			writeError(w, http.StatusNotFound, "cobro no encontrado")
			return
		} else if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudo leer el cobro")
			return
		}

		var c domain.CobroRenta
		if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
			writeError(w, http.StatusBadRequest, "cuerpo de la petición inválido")
			return
		}
		if msg := validateCobroCampos(c); msg != "" {
			writeError(w, http.StatusBadRequest, msg)
			return
		}

		actualizado, err := cobros.Update(id, c)
		if errors.Is(err, sqlite.ErrNotFound) {
			writeError(w, http.StatusNotFound, "cobro no encontrado")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudo actualizar el cobro")
			return
		}
		writeJSON(w, http.StatusOK, actualizado)
	}
}
