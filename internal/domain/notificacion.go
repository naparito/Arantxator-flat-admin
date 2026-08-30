package domain

import (
	"fmt"
	"time"
)

// TipoNotificacion enumera las cuatro reglas del centro de notificaciones
// (§8 del diseño técnico-funcional). No hay tabla de "notificaciones
// generadas": cada aviso se evalúa en caliente cruzando los datos ya
// existentes (contratos, gastos, incidencias) y solo se persiste qué avisos
// ha marcado el usuario como leídos (tabla notificaciones_leidas).
type TipoNotificacion string

const (
	NotifContratoPorVencer  TipoNotificacion = "contrato_por_vencer"
	NotifFianzaSinDepositar TipoNotificacion = "fianza_sin_depositar"
	NotifFacturaPendiente   TipoNotificacion = "factura_pendiente"
	NotifIncidenciaAbierta  TipoNotificacion = "incidencia_abierta"
)

// TiposNotificacion lista los valores válidos, para validar la entrada de la
// API (la `clave` que llega en POST /api/notificaciones/{id}/leida).
var TiposNotificacion = []TipoNotificacion{
	NotifContratoPorVencer, NotifFianzaSinDepositar, NotifFacturaPendiente, NotifIncidenciaAbierta,
}

func (t TipoNotificacion) Valida() bool {
	for _, v := range TiposNotificacion {
		if v == t {
			return true
		}
	}
	return false
}

// SeveridadNotificacion clasifica la urgencia de un aviso, con los mismos
// tres niveles que las pills del mockup (Notificaciones.dc.html).
type SeveridadNotificacion string

const (
	SeveridadUrgente SeveridadNotificacion = "urgente"
	SeveridadAviso   SeveridadNotificacion = "aviso"
	SeveridadInfo    SeveridadNotificacion = "info"
)

var SeveridadesNotificacion = []SeveridadNotificacion{SeveridadUrgente, SeveridadAviso, SeveridadInfo}

func (s SeveridadNotificacion) Valida() bool {
	for _, v := range SeveridadesNotificacion {
		if v == s {
			return true
		}
	}
	return false
}

// Orden da el peso de ordenación de una severidad (menor = más arriba): los
// avisos urgentes primero, luego los avisos normales, luego los informativos.
func (s SeveridadNotificacion) Orden() int {
	switch s {
	case SeveridadUrgente:
		return 0
	case SeveridadAviso:
		return 1
	case SeveridadInfo:
		return 2
	default:
		return 3
	}
}

// DiasFianzaUrgente es el umbral (en días naturales que faltan para el fin
// del plazo legal de depósito de la fianza, DiasPlazoDepositoFianza = 30) por
// debajo del cual el aviso de "fianza sin depositar" pasa de severidad
// "aviso" a "urgente". Fijado en 10 = último tercio del plazo: con 10 días o
// menos de margen —o ya fuera de plazo— queda poco tiempo real de reacción,
// así que el aviso sube a urgente. Con más margen es un "aviso" normal.
const DiasFianzaUrgente = 10

// DiasFacturaAviso es la ventana (en días naturales hasta el vencimiento)
// dentro de la cual una factura todavía "pendiente" genera un aviso. Una
// factura pendiente cuyo vencimiento aún queda lejos —o que no tiene fecha de
// vencimiento— no es accionable todavía y no satura el centro de
// notificaciones. Regla de severidad de "factura pendiente": ya vencida
// (EstadoPagoDerivado == "vencido") -> urgente; pendiente dentro de esta
// ventana -> aviso; pendiente más allá de la ventana -> no genera aviso.
const DiasFacturaAviso = 30

// Notificacion es un aviso del centro de notificaciones, evaluado al leer.
// Clave es su identidad determinista y estable ("<tipo>:<entidadTipo>:<entidadId>"):
// no cambia entre arranques, así que "marcar leída" (POST
// /api/notificaciones/{id}/leida con esa clave) persiste. Fecha es la fecha
// de referencia del aviso (fin del contrato, límite de la fianza,
// vencimiento de la factura, apertura de la incidencia).
type Notificacion struct {
	Clave       string                `json:"clave"`
	Tipo        TipoNotificacion      `json:"tipo"`
	Severidad   SeveridadNotificacion `json:"severidad"`
	Titulo      string                `json:"titulo"`
	Descripcion string                `json:"descripcion"`
	EntidadTipo EntidadTipo           `json:"entidadTipo"`
	EntidadID   int64                 `json:"entidadId"`
	InmuebleID  int64                 `json:"inmuebleId"`
	Fecha       Fecha                 `json:"fecha"`
	Leida       bool                  `json:"leida"`
}

// ClaveNotificacion compone la identidad determinista de un aviso. El mismo
// (tipo, entidad, id) produce siempre la misma clave, de modo que marcarlo
// leído sobrevive a un reinicio y a que la condición desaparezca y vuelva.
func ClaveNotificacion(tipo TipoNotificacion, entidad EntidadTipo, entidadID int64) string {
	return fmt.Sprintf("%s:%s:%d", tipo, entidad, entidadID)
}

// diasNaturales cuenta los días naturales completos entre dos instantes
// (negativo si `hasta` ya pasó respecto a `desde`).
func diasNaturales(desde, hasta time.Time) int {
	return int(soloFecha(hasta).Sub(soloFecha(desde)).Hours() / 24)
}

func fechaCorta(f Fecha) string { return time.Time(f).Format("02/01/2006") }

// textoPlazo verbaliza un desfase en días para las descripciones ("en 12
// días", "mañana", "hoy", "hace 3 días").
func textoPlazo(dias int) string {
	switch {
	case dias > 1:
		return fmt.Sprintf("en %d días", dias)
	case dias == 1:
		return "mañana"
	case dias == 0:
		return "hoy"
	case dias == -1:
		return "ayer"
	default:
		return fmt.Sprintf("hace %d días", -dias)
	}
}

// NotificacionContratoVencimiento genera el aviso de "contrato próximo a
// vencer" cuando el estado derivado del contrato (a la fecha de referencia)
// cae dentro de la ventana de aviso (DiasAvisoVencimiento = 60 días). Un
// contrato ya vencido eleva el aviso a "urgente"; uno rescindido no genera
// nada. Devuelve (_, false) si no procede.
func NotificacionContratoVencimiento(c Contrato, ref time.Time) (Notificacion, bool) {
	estado := c.EstadoDerivado(ref)
	if estado != ContratoProximoAVencer && estado != ContratoVencido {
		return Notificacion{}, false
	}
	sev := SeveridadAviso
	titulo := "Contrato próximo a vencer"
	if estado == ContratoVencido {
		sev = SeveridadUrgente
		titulo = "Contrato vencido"
	}
	dias := diasNaturales(ref, time.Time(c.FechaFin))
	return Notificacion{
		Clave:       ClaveNotificacion(NotifContratoPorVencer, EntidadContrato, c.ID),
		Tipo:        NotifContratoPorVencer,
		Severidad:   sev,
		Titulo:      titulo,
		Descripcion: fmt.Sprintf("Vence el %s (%s).", fechaCorta(c.FechaFin), textoPlazo(dias)),
		EntidadTipo: EntidadContrato,
		EntidadID:   c.ID,
		InmuebleID:  c.InmuebleID,
		Fecha:       c.FechaFin,
	}, true
}

// NotificacionFianza genera el aviso de "fianza sin depositar" mientras la
// fianza siga "pendiente" y el contrato no esté rescindido. La severidad
// depende de cuántos días quedan para la fecha límite de depósito (firma +
// DiasPlazoDepositoFianza): DiasFianzaUrgente o menos -> urgente; más
// margen -> aviso.
func NotificacionFianza(c Contrato, ref time.Time) (Notificacion, bool) {
	if c.Estado == ContratoRescindido || c.FianzaEstado != FianzaPendiente {
		return Notificacion{}, false
	}
	limite := FechaLimiteDepositoFianza(time.Time(c.FechaFirma))
	dias := diasNaturales(ref, limite)
	sev := SeveridadAviso
	if dias <= DiasFianzaUrgente {
		sev = SeveridadUrgente
	}
	return Notificacion{
		Clave:     ClaveNotificacion(NotifFianzaSinDepositar, EntidadContrato, c.ID),
		Tipo:      NotifFianzaSinDepositar,
		Severidad: sev,
		Titulo:    "Fianza sin depositar",
		Descripcion: fmt.Sprintf("Plazo legal de %d días desde la firma. Fecha límite: %s (%s).",
			DiasPlazoDepositoFianza, limite.Format("02/01/2006"), textoPlazo(dias)),
		EntidadTipo: EntidadContrato,
		EntidadID:   c.ID,
		InmuebleID:  c.InmuebleID,
		Fecha:       Fecha(limite),
	}, true
}

// NotificacionFactura genera el aviso de "factura pendiente". Cuenta toda
// factura cuyo estado de pago derivado sea "vencido" (-> urgente) o
// "pendiente" con vencimiento dentro de DiasFacturaAviso (-> aviso). Una
// factura pagada, sin fecha de vencimiento, o con vencimiento aún lejano no
// genera aviso.
func NotificacionFactura(g Gasto, ref time.Time) (Notificacion, bool) {
	estado := g.EstadoPagoDerivado(ref)
	if estado == PagoPagado || g.FechaVencimiento == nil {
		return Notificacion{}, false
	}
	dias := diasNaturales(ref, time.Time(*g.FechaVencimiento))
	var sev SeveridadNotificacion
	var titulo, desc string
	switch estado {
	case PagoVencido:
		sev = SeveridadUrgente
		titulo = fmt.Sprintf("Factura de %s vencida", g.Tipo)
		desc = fmt.Sprintf("%.2f € vencieron el %s (%s).", g.Importe, fechaCorta(*g.FechaVencimiento), textoPlazo(dias))
	default: // pendiente
		if dias > DiasFacturaAviso {
			return Notificacion{}, false
		}
		sev = SeveridadAviso
		titulo = fmt.Sprintf("Factura de %s pendiente", g.Tipo)
		desc = fmt.Sprintf("%.2f € vencen el %s (%s).", g.Importe, fechaCorta(*g.FechaVencimiento), textoPlazo(dias))
	}
	return Notificacion{
		Clave:       ClaveNotificacion(NotifFacturaPendiente, EntidadGasto, g.ID),
		Tipo:        NotifFacturaPendiente,
		Severidad:   sev,
		Titulo:      titulo,
		Descripcion: desc,
		EntidadTipo: EntidadGasto,
		EntidadID:   g.ID,
		InmuebleID:  g.InmuebleID,
		Fecha:       *g.FechaVencimiento,
	}, true
}

// NotificacionIncidencia genera el aviso de "incidencia abierta" mientras la
// incidencia esté en un estado de trabajo pendiente: "abierta", "en proceso"
// o "esperando proveedor". Una incidencia "resuelta" (trabajo hecho, solo
// falta el cierre administrativo) o "cerrada" no genera aviso. Severidad:
// prioridad urgente -> urgente; estado "abierta" -> aviso (aún sin gestionar);
// ya en gestión ("en proceso" / "esperando proveedor") -> info.
func NotificacionIncidencia(i Incidencia, ref time.Time) (Notificacion, bool) {
	switch i.Estado {
	case IncidenciaAbierta, IncidenciaEnProceso, IncidenciaEsperandoProveedor:
		// procede
	default:
		return Notificacion{}, false
	}
	sev := SeveridadInfo
	titulo := "Incidencia en proceso"
	switch i.Estado {
	case IncidenciaAbierta:
		sev = SeveridadAviso
		titulo = "Incidencia abierta"
	case IncidenciaEsperandoProveedor:
		titulo = "Incidencia esperando al proveedor"
	}
	if i.Prioridad == PrioridadUrgente {
		sev = SeveridadUrgente
	}
	desc := fmt.Sprintf("%s · prioridad %s.", i.Titulo, i.Prioridad)
	if i.ProveedorNombre != "" {
		desc = fmt.Sprintf("%s · prioridad %s. %s asignada.", i.Titulo, i.Prioridad, i.ProveedorNombre)
	}
	return Notificacion{
		Clave:       ClaveNotificacion(NotifIncidenciaAbierta, EntidadIncidencia, i.ID),
		Tipo:        NotifIncidenciaAbierta,
		Severidad:   sev,
		Titulo:      titulo,
		Descripcion: desc,
		EntidadTipo: EntidadIncidencia,
		EntidadID:   i.ID,
		InmuebleID:  i.InmuebleID,
		Fecha:       Fecha(soloFecha(i.FechaApertura)),
	}, true
}
