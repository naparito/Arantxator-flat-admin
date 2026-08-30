package httpapi

import (
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/naparito/Arantxator-flat-admin/internal/domain"
	"github.com/naparito/Arantxator-flat-admin/internal/storage/sqlite"
)

// diaUTC descarta la parte horaria (y la zona) de un instante para comparar
// por día natural, igual que sqlite.soloDia en la capa de repositorios.
func diaUTC(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

var entidadesNotificacionValidas = map[domain.EntidadTipo]bool{
	domain.EntidadContrato:   true,
	domain.EntidadGasto:      true,
	domain.EntidadIncidencia: true,
}

// evaluarNotificaciones construye la lista completa de avisos ACTIVOS del
// centro de notificaciones cruzando en caliente los repos ya existentes
// (contratos, gastos, incidencias) — no hay tabla de avisos generados ni
// proceso en segundo plano. Marca Leida = true en los que el usuario ya
// descartó (tabla notificaciones_leidas), enriquece la descripción con la
// dirección del inmueble y ordena por severidad y luego por fecha.
func evaluarNotificaciones(
	ref time.Time,
	inmuebles *sqlite.InmueblesRepo,
	contratos *sqlite.ContratosRepo,
	gastos *sqlite.GastosRepo,
	incidencias *sqlite.IncidenciasRepo,
	notifs *sqlite.NotificacionesRepo,
) ([]domain.Notificacion, error) {
	ms, err := inmuebles.List("")
	if err != nil {
		return nil, err
	}
	direccion := make(map[int64]string, len(ms))
	for _, m := range ms {
		direccion[m.ID] = m.Direccion
	}

	out := []domain.Notificacion{}

	cs, err := contratos.List()
	if err != nil {
		return nil, err
	}
	for _, c := range cs {
		if n, ok := domain.NotificacionContratoVencimiento(c, ref); ok {
			out = append(out, n)
		}
		if n, ok := domain.NotificacionFianza(c, ref); ok {
			out = append(out, n)
		}
	}

	gs, err := gastos.List()
	if err != nil {
		return nil, err
	}
	for _, g := range gs {
		if n, ok := domain.NotificacionFactura(g, ref); ok {
			out = append(out, n)
		}
	}

	is, err := incidencias.List()
	if err != nil {
		return nil, err
	}
	for _, i := range is {
		if n, ok := domain.NotificacionIncidencia(i, ref); ok {
			out = append(out, n)
		}
	}

	leidas, err := notifs.ClavesLeidas()
	if err != nil {
		return nil, err
	}
	for idx := range out {
		if _, ok := leidas[out[idx].Clave]; ok {
			out[idx].Leida = true
		}
		if d := direccion[out[idx].InmuebleID]; d != "" {
			out[idx].Descripcion = d + " — " + out[idx].Descripcion
		}
	}

	sort.SliceStable(out, func(a, b int) bool {
		na, nb := out[a], out[b]
		if na.Severidad.Orden() != nb.Severidad.Orden() {
			return na.Severidad.Orden() < nb.Severidad.Orden()
		}
		return time.Time(na.Fecha).Before(time.Time(nb.Fecha))
	})
	return out, nil
}

type notificacionesResp struct {
	Notificaciones []domain.Notificacion `json:"notificaciones"`
	TotalActivas   int                   `json:"totalActivas"`
	TotalSinLeer   int                   `json:"totalSinLeer"`
}

func handleListNotificaciones(
	inmuebles *sqlite.InmueblesRepo,
	contratos *sqlite.ContratosRepo,
	gastos *sqlite.GastosRepo,
	incidencias *sqlite.IncidenciasRepo,
	notifs *sqlite.NotificacionesRepo,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lista, err := evaluarNotificaciones(time.Now(), inmuebles, contratos, gastos, incidencias, notifs)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudieron evaluar las notificaciones")
			return
		}
		sinLeer := 0
		for _, n := range lista {
			if !n.Leida {
				sinLeer++
			}
		}
		writeJSON(w, http.StatusOK, notificacionesResp{
			Notificaciones: lista,
			TotalActivas:   len(lista),
			TotalSinLeer:   sinLeer,
		})
	}
}

// handleMarcarNotificacionLeida recibe en {id} la `clave` determinista del
// aviso ("<tipo>:<entidadTipo>:<entidadId>") y la guarda en
// notificaciones_leidas. Es idempotente y NO toca el dato subyacente: el
// contrato sigue "próximo a vencer" aunque su aviso quede leído; solo
// desaparece del contador.
func handleMarcarNotificacionLeida(notifs *sqlite.NotificacionesRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clave := r.PathValue("id")
		partes := strings.SplitN(clave, ":", 3)
		if len(partes) != 3 {
			writeError(w, http.StatusBadRequest, "clave de notificación inválida, se espera <tipo>:<entidad>:<id>")
			return
		}
		tipo := domain.TipoNotificacion(partes[0])
		entidad := domain.EntidadTipo(partes[1])
		entidadID, err := strconv.ParseInt(partes[2], 10, 64)
		if err != nil || !tipo.Valida() || !entidadesNotificacionValidas[entidad] {
			writeError(w, http.StatusBadRequest, "clave de notificación inválida")
			return
		}
		if err := notifs.MarcarLeida(clave, tipo, entidad, entidadID); err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudo marcar la notificación como leída")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"clave": clave, "leida": true})
	}
}

func handleGetDashboardResumen(
	inmuebles *sqlite.InmueblesRepo,
	contratos *sqlite.ContratosRepo,
	gastos *sqlite.GastosRepo,
	incidencias *sqlite.IncidenciasRepo,
	cobros *sqlite.CobrosRepo,
	notifs *sqlite.NotificacionesRepo,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ref := time.Now()
		etiqueta, desde, hasta, err := periodoMes(r.URL.Query().Get("periodo"))
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		ms, err := inmuebles.List("")
		if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudieron listar los inmuebles")
			return
		}
		inmueblesOcupados, habitacionesTotales, habitacionesOcupadas := 0, 0, 0
		for _, m := range ms {
			if m.Compartido {
				oc, to, err := contratos.OcupacionInmueble(m.ID, ref)
				if err != nil {
					writeError(w, http.StatusInternalServerError, "no se pudo calcular la ocupación")
					return
				}
				habitacionesTotales += to
				habitacionesOcupadas += oc
				if oc > 0 {
					inmueblesOcupados++
				}
			} else if m.Estado == domain.InmuebleAlquilado {
				inmueblesOcupados++
			}
		}
		porcentajeOcupacion := 0
		if len(ms) > 0 {
			porcentajeOcupacion = int(float64(inmueblesOcupados)/float64(len(ms))*100 + 0.5)
		}

		cs, err := contratos.List()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudieron listar los contratos")
			return
		}
		contratosPorVencer := 0
		for _, c := range cs {
			if c.EstadoDerivado(ref) == domain.ContratoProximoAVencer {
				contratosPorVencer++
			}
		}

		gs, err := gastos.List()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudieron listar los gastos")
			return
		}
		gastosPendientesCant := 0
		gastosPendientesImporte := 0.0
		gastosDelMes := 0.0
		for _, g := range gs {
			if g.EstadoPagoDerivado(ref) != domain.PagoPagado {
				gastosPendientesCant++
				gastosPendientesImporte += g.Importe
			}
			em := diaUTC(time.Time(g.FechaEmision))
			if !em.Before(desde) && em.Before(hasta) {
				gastosDelMes += g.Importe
			}
		}

		is, err := incidencias.List()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudieron listar las incidencias")
			return
		}
		incidenciasAbiertas := 0
		for _, i := range is {
			if !i.Estado.EsFinal() {
				incidenciasAbiertas++
			}
		}

		cbs, err := cobros.List()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudieron listar los cobros")
			return
		}
		ingresosDelMes := 0.0
		for _, c := range cbs {
			p := diaUTC(time.Time(c.Periodo))
			if !p.Before(desde) && p.Before(hasta) {
				ingresosDelMes += c.Importe
			}
		}

		lista, err := evaluarNotificaciones(ref, inmuebles, contratos, gastos, incidencias, notifs)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudieron evaluar las notificaciones")
			return
		}
		notificacionesSinLeer := 0
		for _, n := range lista {
			if !n.Leida {
				notificacionesSinLeer++
			}
		}

		writeJSON(w, http.StatusOK, domain.ResumenDashboard{
			Periodo: etiqueta,
			Ocupacion: domain.OcupacionCartera{
				InmueblesTotales:     len(ms),
				InmueblesOcupados:    inmueblesOcupados,
				HabitacionesTotales:  habitacionesTotales,
				HabitacionesOcupadas: habitacionesOcupadas,
				Porcentaje:           porcentajeOcupacion,
			},
			ContratosPorVencer: contratosPorVencer,
			GastosPendientes: domain.GastosPendientesResumen{
				Cantidad: gastosPendientesCant,
				Importe:  redondea2(gastosPendientesImporte),
			},
			IncidenciasAbiertas: incidenciasAbiertas,
			Rentabilidad: domain.RentabilidadResumen{
				Periodo:  etiqueta,
				Ingresos: redondea2(ingresosDelMes),
				Gastos:   redondea2(gastosDelMes),
				Neto:     redondea2(ingresosDelMes - gastosDelMes),
			},
			NotificacionesSinLeer: notificacionesSinLeer,
		})
	}
}
