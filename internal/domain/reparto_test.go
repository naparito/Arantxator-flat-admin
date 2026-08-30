package domain

import (
	"math"
	"testing"
	"time"
)

func fecha(y int, m time.Month, d int) Fecha {
	return Fecha(time.Date(y, m, d, 0, 0, 0, 0, time.UTC))
}

func cuota(inquilinoID int64, pct float64) RepartoGasto {
	return RepartoGasto{InquilinoID: inquilinoID, Porcentaje: pct}
}

func TestCalcularRecibo_CasoDeReferencia78Euros(t *testing.T) {
	// El caso literal del plan: 78,00 € a 33/33/34 % debe dar
	// 25,74 + 25,74 + 26,52 y sumar 78,00 € exacto (verificar el redondeo).
	g := Gasto{ID: 1, Tipo: GastoLuz, Importe: 78.00, FechaEmision: fecha(2026, 9, 1)}
	cuotas := []RepartoGasto{cuota(1, 33), cuota(2, 33), cuota(3, 34)}

	rec := CalcularRecibo(g, cuotas)
	if rec.SinReparto {
		t.Fatal("con reparto vigente, SinReparto debería ser false")
	}
	if len(rec.Lineas) != 3 {
		t.Fatalf("esperaba 3 líneas de recibo, obtuve %d", len(rec.Lineas))
	}
	quiero := []float64{25.74, 25.74, 26.52}
	suma := 0.0
	for i, l := range rec.Lineas {
		if math.Abs(l.Importe-quiero[i]) > 1e-9 {
			t.Fatalf("línea %d: esperaba %.2f €, obtuve %.2f €", i, quiero[i], l.Importe)
		}
		suma += l.Importe
	}
	if math.Abs(suma-78.00) > 1e-9 {
		t.Fatalf("la suma de los recibos (%.2f) no cuadra con el total 78,00 €", suma)
	}
}

func TestCalcularRecibo_LaSumaSiempreCuadraConElTotal(t *testing.T) {
	casos := []struct {
		nombre  string
		importe float64
		pcts    []float64
	}{
		{"tercios exactos decimales", 100.00, []float64{33.33, 33.33, 33.34}},
		{"importe con céntimos feos", 87.31, []float64{33, 33, 34}},
		{"dos inquilinos 50/50", 45.05, []float64{50, 50}},
		{"reparto desigual", 199.99, []float64{10, 25, 65}},
		{"cuatro inquilinos", 1000.01, []float64{25, 25, 25, 25}},
		{"un solo inquilino al 100", 63.40, []float64{100}},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			cuotas := make([]RepartoGasto, len(c.pcts))
			for i, p := range c.pcts {
				cuotas[i] = cuota(int64(i+1), p)
			}
			rec := CalcularRecibo(Gasto{Importe: c.importe}, cuotas)
			totalCent := int64(math.Round(c.importe * 100))
			var sumaCent int64
			for _, l := range rec.Lineas {
				sumaCent += int64(math.Round(l.Importe * 100))
			}
			if sumaCent != totalCent {
				t.Fatalf("la suma de las líneas (%d cént.) no cuadra con el total (%d cént.)", sumaCent, totalCent)
			}
		})
	}
}

func TestCalcularRecibo_SinRepartoVigente(t *testing.T) {
	// Gasto en un inmueble sin reparto (no compartido, o compartido sin
	// reparto para ese tipo): no es un error, es un recibo sin líneas.
	rec := CalcularRecibo(Gasto{ID: 7, Importe: 50}, nil)
	if !rec.SinReparto {
		t.Fatal("sin reparto vigente, SinReparto debería ser true")
	}
	if len(rec.Lineas) != 0 {
		t.Fatalf("sin reparto no debería haber líneas, obtuve %d", len(rec.Lineas))
	}
}

func TestRepartoVigenteEnFecha_EligeLaVersionDeLaEpoca(t *testing.T) {
	// Dos versiones del reparto de "luz": la vieja [.., 2026-03-01) y la
	// nueva [2026-03-01, ..). Intervalos semiabiertos.
	hasta := fecha(2026, 3, 1)
	reparto := []RepartoGasto{
		{InquilinoID: 1, TipoGasto: GastoLuz, Porcentaje: 50, VigenteDesde: fecha(2026, 1, 1), VigenteHasta: &hasta},
		{InquilinoID: 2, TipoGasto: GastoLuz, Porcentaje: 50, VigenteDesde: fecha(2026, 1, 1), VigenteHasta: &hasta},
		{InquilinoID: 1, TipoGasto: GastoLuz, Porcentaje: 33, VigenteDesde: fecha(2026, 3, 1)},
		{InquilinoID: 2, TipoGasto: GastoLuz, Porcentaje: 33, VigenteDesde: fecha(2026, 3, 1)},
		{InquilinoID: 3, TipoGasto: GastoLuz, Porcentaje: 34, VigenteDesde: fecha(2026, 3, 1)},
		{InquilinoID: 1, TipoGasto: GastoAgua, Porcentaje: 100, VigenteDesde: fecha(2026, 1, 1)},
	}

	antes := RepartoVigenteEnFecha(reparto, GastoLuz, time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC))
	if len(antes) != 2 || antes[0].Porcentaje != 50 {
		t.Fatalf("una factura de febrero debería usar el reparto viejo (2 al 50%%), obtuve %+v", antes)
	}

	justoEnElCambio := RepartoVigenteEnFecha(reparto, GastoLuz, time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC))
	if len(justoEnElCambio) != 3 {
		t.Fatalf("una factura del 01/03 (inicio de vigencia) ya usa el reparto nuevo, obtuve %+v", justoEnElCambio)
	}

	despues := RepartoVigenteEnFecha(reparto, GastoLuz, time.Date(2026, 9, 20, 0, 0, 0, 0, time.UTC))
	if len(despues) != 3 || despues[2].Porcentaje != 34 {
		t.Fatalf("una factura de septiembre debería usar el reparto nuevo (33/33/34), obtuve %+v", despues)
	}

	// Un tipo de gasto sin reparto definido -> sin cuotas, sin pánico.
	if got := RepartoVigenteEnFecha(reparto, GastoGas, time.Now()); len(got) != 0 {
		t.Fatalf("gas no tiene reparto, esperaba 0 cuotas, obtuve %+v", got)
	}
}

func TestSumaPorcentajesCuadra(t *testing.T) {
	if !SumaPorcentajesCuadra([]float64{33, 33, 34}) {
		t.Fatal("33+33+34 = 100 debería cuadrar")
	}
	if !SumaPorcentajesCuadra([]float64{33.33, 33.33, 33.34}) {
		t.Fatal("33.33+33.33+33.34 = 100 debería cuadrar (con tolerancia de redondeo)")
	}
	if SumaPorcentajesCuadra([]float64{50, 40}) {
		t.Fatal("50+40 = 90 NO debería cuadrar")
	}
	if SumaPorcentajesCuadra([]float64{60, 60}) {
		t.Fatal("60+60 = 120 NO debería cuadrar")
	}
}
