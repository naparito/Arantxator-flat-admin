package domain

import "testing"

func TestTransicionEstadoIncidenciaValida_FlujoFeliz(t *testing.T) {
	pasos := []struct{ desde, hasta EstadoIncidencia }{
		{IncidenciaAbierta, IncidenciaEnProceso},
		{IncidenciaEnProceso, IncidenciaEsperandoProveedor},
		{IncidenciaEsperandoProveedor, IncidenciaResuelta},
		{IncidenciaResuelta, IncidenciaCerrada},
	}
	for _, p := range pasos {
		if !TransicionEstadoIncidenciaValida(p.desde, p.hasta) {
			t.Errorf("%s -> %s debería ser una transición válida del flujo", p.desde, p.hasta)
		}
	}
}

func TestTransicionEstadoIncidenciaValida_Rechazos(t *testing.T) {
	casos := []struct {
		nombre       string
		desde, hasta EstadoIncidencia
	}{
		{"salto de dos pasos", IncidenciaAbierta, IncidenciaEsperandoProveedor},
		{"retroceso de un paso que no es reapertura", IncidenciaEsperandoProveedor, IncidenciaEnProceso},
		{"mismo estado", IncidenciaEnProceso, IncidenciaEnProceso},
		{"estado destino inválido", IncidenciaAbierta, EstadoIncidencia("pendiente")},
		{"estado origen inválido", EstadoIncidencia(""), IncidenciaEnProceso},
	}
	for _, c := range casos {
		if TransicionEstadoIncidenciaValida(c.desde, c.hasta) {
			t.Errorf("%s: %s -> %s NO debería permitirse", c.nombre, c.desde, c.hasta)
		}
	}
}

func TestTransicionEstadoIncidenciaValida_Reapertura(t *testing.T) {
	if !TransicionEstadoIncidenciaValida(IncidenciaResuelta, IncidenciaEnProceso) {
		t.Error("una incidencia resuelta debería poder reabrirse a 'en_proceso'")
	}
	if !TransicionEstadoIncidenciaValida(IncidenciaCerrada, IncidenciaEnProceso) {
		t.Error("una incidencia cerrada debería poder reabrirse a 'en_proceso'")
	}
}

func TestEstadoIncidenciaEsFinal(t *testing.T) {
	if !IncidenciaCerrada.EsFinal() {
		t.Error("'cerrada' debería contar como estado final")
	}
	for _, e := range []EstadoIncidencia{IncidenciaAbierta, IncidenciaEnProceso, IncidenciaEsperandoProveedor, IncidenciaResuelta} {
		if e.EsFinal() {
			t.Errorf("%s no debería contar como estado final (sigue en el contador de abiertas)", e)
		}
	}
}
