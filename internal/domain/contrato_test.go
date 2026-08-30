package domain

import (
	"testing"
	"time"
)

func TestDuracionMinimaLAUAnios(t *testing.T) {
	if got := DuracionMinimaLAUAnios(false); got != 5 {
		t.Fatalf("arrendador persona física: esperaba 5 años, obtuve %d", got)
	}
	if got := DuracionMinimaLAUAnios(true); got != 7 {
		t.Fatalf("arrendador persona jurídica: esperaba 7 años, obtuve %d", got)
	}
}

func TestFechaFinSugeridaLAU(t *testing.T) {
	firma := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

	if got := FechaFinSugeridaLAU(firma, false); !got.Equal(time.Date(2031, 2, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("persona física: esperaba 2031-02-01, obtuve %s", got.Format("2006-01-02"))
	}
	if got := FechaFinSugeridaLAU(firma, true); !got.Equal(time.Date(2033, 2, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("persona jurídica: esperaba 2033-02-01, obtuve %s", got.Format("2006-01-02"))
	}
}

func TestFechaLimiteDepositoFianza(t *testing.T) {
	// La fecha límite es exactamente firma + 30 días.
	firma := time.Date(2026, 7, 30, 0, 0, 0, 0, time.UTC)
	got := FechaLimiteDepositoFianza(firma)
	want := time.Date(2026, 8, 29, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("esperaba %s (firma + 30 días), obtuve %s", want.Format("2006-01-02"), got.Format("2006-01-02"))
	}
}

func TestContratoEstadoDerivado(t *testing.T) {
	ref := time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC)
	fecha := func(y int, m time.Month, d int) Fecha {
		return Fecha(time.Date(y, m, d, 0, 0, 0, 0, time.UTC))
	}

	casos := []struct {
		nombre   string
		contrato Contrato
		quiero   EstadoContrato
	}{
		{
			nombre:   "fin lejano -> activo",
			contrato: Contrato{FechaFin: fecha(2030, 1, 31)},
			quiero:   ContratoActivo,
		},
		{
			nombre:   "fin dentro de la ventana de 60 días -> próximo a vencer",
			contrato: Contrato{FechaFin: fecha(2026, 10, 15)}, // ~46 días
			quiero:   ContratoProximoAVencer,
		},
		{
			nombre:   "fin justo en el borde de la ventana (60 días) -> próximo a vencer",
			contrato: Contrato{FechaFin: fecha(2026, 10, 29)},
			quiero:   ContratoProximoAVencer,
		},
		{
			nombre:   "fin a 61 días -> todavía activo",
			contrato: Contrato{FechaFin: fecha(2026, 10, 30)},
			quiero:   ContratoActivo,
		},
		{
			nombre:   "fin ya pasado -> vencido",
			contrato: Contrato{FechaFin: fecha(2026, 8, 1)},
			quiero:   ContratoVencido,
		},
		{
			nombre:   "rescindido a mano -> se conserva, no se recalcula",
			contrato: Contrato{FechaFin: fecha(2030, 1, 1), Estado: ContratoRescindido},
			quiero:   ContratoRescindido,
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if got := c.contrato.EstadoDerivado(ref); got != c.quiero {
				t.Fatalf("esperaba %q, obtuve %q", c.quiero, got)
			}
		})
	}
}

func TestContratoEstaVigente(t *testing.T) {
	ref := time.Date(2026, 8, 30, 0, 0, 0, 0, time.UTC)
	fecha := func(y int, m time.Month, d int) Fecha { return Fecha(time.Date(y, m, d, 0, 0, 0, 0, time.UTC)) }

	if (Contrato{FechaFin: fecha(2030, 1, 1)}).EstaVigente(ref) != true {
		t.Fatal("un contrato activo debería estar vigente")
	}
	if (Contrato{FechaFin: fecha(2026, 9, 15)}).EstaVigente(ref) != true {
		t.Fatal("un contrato próximo a vencer sigue vigente (ocupa su ámbito)")
	}
	if (Contrato{FechaFin: fecha(2026, 1, 1)}).EstaVigente(ref) != false {
		t.Fatal("un contrato vencido no está vigente")
	}
	if (Contrato{FechaFin: fecha(2030, 1, 1), Estado: ContratoRescindido}).EstaVigente(ref) != false {
		t.Fatal("un contrato rescindido no está vigente")
	}
}
