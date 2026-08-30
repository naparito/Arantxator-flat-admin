package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/naparito/Arantxator-flat-admin/internal/domain"
	"github.com/naparito/Arantxator-flat-admin/internal/storage/sqlite"
)

// cuotaEntrada es una línea de reparto tal como la manda la GUI.
type cuotaEntrada struct {
	InquilinoID int64   `json:"inquilinoId"`
	Porcentaje  float64 `json:"porcentaje"`
}

// repartoEntrada es el cuerpo de POST /api/inmuebles/{id}/reparto: una
// versión nueva del reparto de un tipo de gasto.
type repartoEntrada struct {
	TipoGasto    domain.TipoGasto `json:"tipoGasto"`
	VigenteDesde domain.Fecha     `json:"vigenteDesde"`
	Motivo       string           `json:"motivo"`
	Cuotas       []cuotaEntrada   `json:"cuotas"`
}

// versionReparto agrupa las filas de una misma versión (un mismo
// inmueble_id + tipo_gasto + vigente_desde) para la respuesta de la API: la
// GUI no tiene que reconstruir la matriz cruzando filas sueltas.
type versionReparto struct {
	TipoGasto    domain.TipoGasto `json:"tipoGasto"`
	VigenteDesde domain.Fecha     `json:"vigenteDesde"`
	VigenteHasta *domain.Fecha    `json:"vigenteHasta"`
	Motivo       string           `json:"motivo"`
	Vigente      bool             `json:"vigente"` // cubre la fecha de hoy
	Cuotas       []cuotaEntrada   `json:"cuotas"`
}

type repartoRespuesta struct {
	InmuebleID int64            `json:"inmuebleId"`
	Versiones  []versionReparto `json:"versiones"`
}

func agruparVersiones(filas []domain.RepartoGasto, ref time.Time) []versionReparto {
	type clave struct {
		tipo  domain.TipoGasto
		desde string
	}
	orden := []clave{}
	porClave := map[clave]*versionReparto{}
	dia := ref.UTC().Format("2006-01-02")

	for _, f := range filas {
		k := clave{f.TipoGasto, time.Time(f.VigenteDesde).Format("2006-01-02")}
		v, ok := porClave[k]
		if !ok {
			desdeStr := time.Time(f.VigenteDesde).Format("2006-01-02")
			hasta := ""
			if f.VigenteHasta != nil {
				hasta = time.Time(*f.VigenteHasta).Format("2006-01-02")
			}
			v = &versionReparto{
				TipoGasto:    f.TipoGasto,
				VigenteDesde: f.VigenteDesde,
				VigenteHasta: f.VigenteHasta,
				Motivo:       f.Motivo,
				Vigente:      desdeStr <= dia && (hasta == "" || dia < hasta),
				Cuotas:       []cuotaEntrada{},
			}
			porClave[k] = v
			orden = append(orden, k)
		}
		v.Cuotas = append(v.Cuotas, cuotaEntrada{InquilinoID: f.InquilinoID, Porcentaje: f.Porcentaje})
	}

	out := make([]versionReparto, 0, len(orden))
	for _, k := range orden {
		out = append(out, *porClave[k])
	}
	// Tipo de gasto, y dentro de cada tipo la vigencia más reciente primero.
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].TipoGasto != out[j].TipoGasto {
			return out[i].TipoGasto < out[j].TipoGasto
		}
		return time.Time(out[i].VigenteDesde).After(time.Time(out[j].VigenteDesde))
	})
	return out
}

func handleGetRepartoInmueble(inmuebles *sqlite.InmueblesRepo, repartos *sqlite.RepartosRepo) http.HandlerFunc {
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

		filas, err := repartos.ListByInmueble(inmuebleID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudo leer el reparto")
			return
		}
		writeJSON(w, http.StatusOK, repartoRespuesta{
			InmuebleID: inmuebleID,
			Versiones:  agruparVersiones(filas, time.Now()),
		})
	}
}

func handleCreateRepartoInmueble(inmuebles *sqlite.InmueblesRepo, repartos *sqlite.RepartosRepo, inquilinos *sqlite.InquilinosRepo) http.HandlerFunc {
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

		var in repartoEntrada
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			writeError(w, http.StatusBadRequest, "cuerpo de la petición inválido")
			return
		}
		if !in.TipoGasto.Valida() {
			writeError(w, http.StatusBadRequest, "tipo de gasto no válido: "+string(in.TipoGasto))
			return
		}
		if time.Time(in.VigenteDesde).IsZero() {
			writeError(w, http.StatusBadRequest, "el campo vigenteDesde es obligatorio")
			return
		}
		if len(in.Cuotas) == 0 {
			writeError(w, http.StatusBadRequest, "el reparto necesita al menos una cuota")
			return
		}

		porcentajes := make([]float64, 0, len(in.Cuotas))
		vistos := map[int64]bool{}
		for _, c := range in.Cuotas {
			if vistos[c.InquilinoID] {
				writeError(w, http.StatusBadRequest, fmt.Sprintf("el inquilino %d aparece dos veces en el reparto", c.InquilinoID))
				return
			}
			vistos[c.InquilinoID] = true
			if c.Porcentaje < 0 || c.Porcentaje > 100 {
				writeError(w, http.StatusBadRequest, "cada porcentaje debe estar entre 0 y 100")
				return
			}
			if _, err := inquilinos.Get(c.InquilinoID); errors.Is(err, sqlite.ErrNotFound) {
				writeError(w, http.StatusBadRequest, fmt.Sprintf("el inquilino %d no existe", c.InquilinoID))
				return
			} else if err != nil {
				writeError(w, http.StatusInternalServerError, "no se pudo comprobar el inquilino")
				return
			}
			porcentajes = append(porcentajes, c.Porcentaje)
		}

		// Regla entre filas: los porcentajes de un tipo de gasto en una misma
		// vigencia deben sumar 100. No es un CHECK de columna -> 400 aquí, y
		// no se guarda nada a medias.
		if !domain.SumaPorcentajesCuadra(porcentajes) {
			writeError(w, http.StatusBadRequest, "los porcentajes del reparto deben sumar 100%")
			return
		}

		cuotas := make([]sqlite.Cuota, len(in.Cuotas))
		for i, c := range in.Cuotas {
			cuotas[i] = sqlite.Cuota{InquilinoID: c.InquilinoID, Porcentaje: c.Porcentaje}
		}

		if _, err := repartos.CrearVersion(inmuebleID, in.TipoGasto, in.VigenteDesde, in.Motivo, cuotas); err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudo guardar el reparto")
			return
		}

		// Devolver el reparto completo del inmueble, ya con la versión nueva.
		todas, err := repartos.ListByInmueble(inmuebleID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudo releer el reparto")
			return
		}
		writeJSON(w, http.StatusCreated, repartoRespuesta{
			InmuebleID: inmuebleID,
			Versiones:  agruparVersiones(todas, time.Now()),
		})
	}
}
