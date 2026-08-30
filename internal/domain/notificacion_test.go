package domain

import (
	"testing"
	"time"
)

// refNotif: fecha de referencia fija para todos los casos ("hoy" = 2026-08-30).
// `fecha(y, m, d)` es el helper de reparto_test.go (mismo paquete).
var refNotif = time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC)

func TestTipoNotificacionValida(t *testing.T) {
	for _, ok := range TiposNotificacion {
		if !ok.Valida() {
			t.Errorf("%q debería ser un tipo válido", ok)
		}
	}
	if TipoNotificacion("cumpleaños").Valida() {
		t.Error("un tipo inventado no debería validar")
	}
}

func TestSeveridadNotificacionValidaYOrden(t *testing.T) {
	if !SeveridadUrgente.Valida() || !SeveridadAviso.Valida() || !SeveridadInfo.Valida() {
		t.Fatal("las tres severidades canónicas deberían validar")
	}
	if SeveridadNotificacion("crítico").Valida() {
		t.Error("una severidad inventada no debería validar")
	}
	if !(SeveridadUrgente.Orden() < SeveridadAviso.Orden() && SeveridadAviso.Orden() < SeveridadInfo.Orden()) {
		t.Errorf("orden de severidades inesperado: urgente=%d aviso=%d info=%d",
			SeveridadUrgente.Orden(), SeveridadAviso.Orden(), SeveridadInfo.Orden())
	}
}

func TestClaveNotificacionEsDeterministaYEstable(t *testing.T) {
	a := ClaveNotificacion(NotifContratoPorVencer, EntidadContrato, 5)
	b := ClaveNotificacion(NotifContratoPorVencer, EntidadContrato, 5)
	if a != b || a != "contrato_por_vencer:contrato:5" {
		t.Fatalf("clave inesperada o no determinista: %q vs %q", a, b)
	}
	// El mismo contrato pero otra regla -> otra clave (fianza y vencimiento
	// conviven sobre el mismo contrato_id).
	if ClaveNotificacion(NotifFianzaSinDepositar, EntidadContrato, 5) == a {
		t.Fatal("dos tipos distintos sobre la misma entidad deberían dar claves distintas")
	}
}

func TestNotificacionContratoVencimiento(t *testing.T) {
	base := Contrato{ID: 7, InmuebleID: 3, FechaFin: fecha(2030, 1, 1)}

	if _, ok := NotificacionContratoVencimiento(base, refNotif); ok {
		t.Fatal("un contrato con fin lejano (año 2030) no debería generar aviso")
	}

	// Fin dentro de la ventana de aviso (60 días): 2026-10-15 ≈ 46 días.
	c := base
	c.FechaFin = fecha(2026, 10, 15)
	n, ok := NotificacionContratoVencimiento(c, refNotif)
	if !ok {
		t.Fatal("un contrato a 46 días debería generar aviso")
	}
	if n.Tipo != NotifContratoPorVencer || n.Severidad != SeveridadAviso {
		t.Fatalf("esperaba tipo contrato_por_vencer / severidad aviso, obtuve %q / %q", n.Tipo, n.Severidad)
	}
	if n.Clave != "contrato_por_vencer:contrato:7" || n.EntidadTipo != EntidadContrato || n.EntidadID != 7 || n.InmuebleID != 3 {
		t.Fatalf("metadatos de la notificación inesperados: %+v", n)
	}

	// Fin ya pasado -> urgente.
	c.FechaFin = fecha(2026, 8, 1)
	n, ok = NotificacionContratoVencimiento(c, refNotif)
	if !ok || n.Severidad != SeveridadUrgente {
		t.Fatalf("un contrato vencido debería dar un aviso urgente, obtuve ok=%v sev=%q", ok, n.Severidad)
	}

	// Rescindido a mano -> nunca genera aviso, aunque la fecha esté cerca.
	c.FechaFin = fecha(2026, 9, 10)
	c.Estado = ContratoRescindido
	if _, ok := NotificacionContratoVencimiento(c, refNotif); ok {
		t.Fatal("un contrato rescindido no debería generar aviso de vencimiento")
	}
}

func TestNotificacionFianza_UrgenteConPocoPlazo_AvisoConMargen(t *testing.T) {
	// Firma hace 26 días -> fecha límite = firma + 30 = dentro de 4 días -> urgente.
	urg := Contrato{ID: 1, InmuebleID: 1, FechaFirma: fecha(2026, 8, 4), FianzaEstado: FianzaPendiente}
	n, ok := NotificacionFianza(urg, refNotif)
	if !ok {
		t.Fatal("una fianza pendiente con el plazo casi agotado debería generar aviso")
	}
	if n.Tipo != NotifFianzaSinDepositar || n.Severidad != SeveridadUrgente {
		t.Fatalf("esperaba fianza_sin_depositar / urgente, obtuve %q / %q", n.Tipo, n.Severidad)
	}
	if n.Clave != "fianza_sin_depositar:contrato:1" {
		t.Fatalf("clave inesperada: %q", n.Clave)
	}

	// Firma hoy -> fecha límite dentro de 30 días -> aviso (hay margen).
	avi := Contrato{ID: 2, InmuebleID: 1, FechaFirma: fecha(2026, 8, 30), FianzaEstado: FianzaPendiente}
	n, ok = NotificacionFianza(avi, refNotif)
	if !ok || n.Severidad != SeveridadAviso {
		t.Fatalf("una fianza recién firmada debería dar un aviso normal, obtuve ok=%v sev=%q", ok, n.Severidad)
	}

	// Justo en el umbral: límite a exactamente DiasFianzaUrgente días -> urgente.
	borde := Contrato{ID: 3, FechaFirma: fecha(2026, 8, 30-30+DiasFianzaUrgente), FianzaEstado: FianzaPendiente}
	if n, ok := NotificacionFianza(borde, refNotif); !ok || n.Severidad != SeveridadUrgente {
		t.Fatalf("con el límite a exactamente %d días esperaba urgente, obtuve ok=%v sev=%q", DiasFianzaUrgente, ok, n.Severidad)
	}

	// Fianza ya depositada -> sin aviso.
	dep := Contrato{ID: 4, FechaFirma: fecha(2026, 8, 4), FianzaEstado: FianzaDepositada}
	if _, ok := NotificacionFianza(dep, refNotif); ok {
		t.Fatal("una fianza ya depositada no debería generar aviso")
	}
}

func TestNotificacionFactura(t *testing.T) {
	venc := func(y int, m time.Month, d int) *Fecha { f := fecha(y, m, d); return &f }

	// Pendiente, vence dentro de la ventana (2026-09-10 ≈ 11 días) -> aviso.
	g := Gasto{ID: 9, InmuebleID: 2, Tipo: GastoComunidad, Importe: 96, EstadoPago: PagoPendiente, FechaVencimiento: venc(2026, 9, 10)}
	n, ok := NotificacionFactura(g, refNotif)
	if !ok || n.Tipo != NotifFacturaPendiente || n.Severidad != SeveridadAviso {
		t.Fatalf("esperaba factura_pendiente / aviso, obtuve ok=%v %q / %q", ok, n.Tipo, n.Severidad)
	}
	if n.Clave != "factura_pendiente:gasto:9" || n.EntidadTipo != EntidadGasto {
		t.Fatalf("metadatos inesperados: %+v", n)
	}

	// Pendiente pero el vencimiento aún queda lejos (> DiasFacturaAviso) -> sin aviso.
	g.FechaVencimiento = venc(2026, 12, 31)
	if _, ok := NotificacionFactura(g, refNotif); ok {
		t.Fatal("una factura pendiente que vence dentro de mucho no debería saturar el centro de avisos")
	}

	// Ya vencida -> urgente.
	g.FechaVencimiento = venc(2026, 8, 20)
	if n, ok := NotificacionFactura(g, refNotif); !ok || n.Severidad != SeveridadUrgente {
		t.Fatalf("una factura vencida debería dar aviso urgente, obtuve ok=%v sev=%q", ok, n.Severidad)
	}

	// Pagada -> sin aviso.
	g.EstadoPago = PagoPagado
	if _, ok := NotificacionFactura(g, refNotif); ok {
		t.Fatal("una factura pagada no debería generar aviso")
	}

	// Sin fecha de vencimiento -> sin aviso (no es accionable).
	g2 := Gasto{ID: 10, Tipo: GastoOtros, Importe: 20, EstadoPago: PagoPendiente}
	if _, ok := NotificacionFactura(g2, refNotif); ok {
		t.Fatal("una factura sin fecha de vencimiento no debería generar aviso")
	}
}

func TestNotificacionIncidencia(t *testing.T) {
	base := Incidencia{ID: 4, InmuebleID: 6, Titulo: "Fuga en el grifo", Prioridad: PrioridadMedia, FechaApertura: refNotif}

	// Abierta -> aviso.
	ab := base
	ab.Estado = IncidenciaAbierta
	if n, ok := NotificacionIncidencia(ab, refNotif); !ok || n.Severidad != SeveridadAviso || n.Tipo != NotifIncidenciaAbierta {
		t.Fatalf("una incidencia abierta debería dar aviso, obtuve ok=%v %q / %q", ok, n.Tipo, n.Severidad)
	}

	// En proceso, prioridad media -> info.
	enp := base
	enp.Estado = IncidenciaEnProceso
	if n, ok := NotificacionIncidencia(enp, refNotif); !ok || n.Severidad != SeveridadInfo {
		t.Fatalf("una incidencia en proceso (prioridad media) debería dar info, obtuve ok=%v sev=%q", ok, n.Severidad)
	}

	// Prioridad urgente -> urgente, sea cual sea el estado no final.
	urg := base
	urg.Estado = IncidenciaEsperandoProveedor
	urg.Prioridad = PrioridadUrgente
	if n, ok := NotificacionIncidencia(urg, refNotif); !ok || n.Severidad != SeveridadUrgente {
		t.Fatalf("una incidencia de prioridad urgente debería dar aviso urgente, obtuve ok=%v sev=%q", ok, n.Severidad)
	}

	// Resuelta (trabajo hecho, solo falta cerrar) -> sin aviso.
	res := base
	res.Estado = IncidenciaResuelta
	if _, ok := NotificacionIncidencia(res, refNotif); ok {
		t.Fatal("una incidencia resuelta no debería generar aviso")
	}

	// Cerrada -> sin aviso.
	cer := base
	cer.Estado = IncidenciaCerrada
	if _, ok := NotificacionIncidencia(cer, refNotif); ok {
		t.Fatal("una incidencia cerrada no debería generar aviso")
	}
}
