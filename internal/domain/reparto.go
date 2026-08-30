package domain

import (
	"math"
	"sort"
	"time"
)

// PorcentajeRepartoTotal es la suma exacta a la que deben llegar los
// porcentajes de un mismo tipo de gasto en una misma versión del reparto.
// Es una regla entre filas (no un CHECK de columna), así que se valida a
// nivel de aplicación: un reparto que no suma 100 no se guarda.
const PorcentajeRepartoTotal = 100.0

// toleranciaPorcentaje absorbe el error de coma flotante al sumar los
// porcentajes tecleados en la GUI (33.33 + 33.33 + 33.34).
const toleranciaPorcentaje = 0.001

// SumaPorcentajesCuadra indica si los porcentajes suman 100 (con tolerancia
// para el redondeo decimal de la entrada).
func SumaPorcentajesCuadra(porcentajes []float64) bool {
	suma := 0.0
	for _, p := range porcentajes {
		suma += p
	}
	return math.Abs(suma-PorcentajeRepartoTotal) <= toleranciaPorcentaje
}

// RepartoVigenteEnFecha filtra, de todas las filas de reparto de un inmueble,
// las de un tipo de gasto concreto cuya vigencia cubre la fecha dada
// (vigente_desde <= fecha < vigente_hasta; vigente_hasta NULL = abierta).
// Devuelve la versión completa (una fila por inquilino), ordenada por
// inquilino para que el reparto sea determinista.
func RepartoVigenteEnFecha(reparto []RepartoGasto, tipo TipoGasto, fecha time.Time) []RepartoGasto {
	dia := soloFecha(fecha)
	vigentes := []RepartoGasto{}
	for _, r := range reparto {
		if r.TipoGasto != tipo {
			continue
		}
		desde := soloFecha(time.Time(r.VigenteDesde))
		if desde.After(dia) {
			continue
		}
		if r.VigenteHasta != nil {
			hasta := soloFecha(time.Time(*r.VigenteHasta))
			if !dia.Before(hasta) {
				continue
			}
		}
		vigentes = append(vigentes, r)
	}
	sort.Slice(vigentes, func(i, j int) bool { return vigentes[i].InquilinoID < vigentes[j].InquilinoID })
	return vigentes
}

// LineaRecibo es lo que le toca pagar a un inquilino de una factura concreta.
type LineaRecibo struct {
	InquilinoID int64   `json:"inquilinoId"`
	Porcentaje  float64 `json:"porcentaje"`
	Importe     float64 `json:"importe"`
}

// Recibo es el desglose por inquilino de una factura de un piso compartido.
// La suma de LineaRecibo.Importe cuadra exactamente con Total (el redondeo
// se reparte, no se pierde ni se inventa un céntimo).
type Recibo struct {
	GastoID    int64         `json:"gastoId"`
	Tipo       TipoGasto     `json:"tipo"`
	Fecha      Fecha         `json:"fecha"`
	Total      float64       `json:"total"`
	Lineas     []LineaRecibo `json:"lineas"`
	SinReparto bool          `json:"sinReparto"`
}

// CalcularRecibo reparte el importe de un gasto entre los inquilinos de su
// reparto vigente. El redondeo se hace en céntimos y el último inquilino
// (por id) absorbe el resto, de modo que la suma de las líneas es siempre
// igual al total al céntimo — caso de referencia del plan: 78,00 € a
// 33/33/34 % → 25,74 + 25,74 + 26,52 = 78,00 € exacto.
//
// Si no hay reparto vigente (inmueble no compartido, o compartido sin
// reparto configurado para ese tipo de gasto) devuelve un recibo sin líneas
// y SinReparto = true: es un caso válido, no un error.
func CalcularRecibo(g Gasto, cuotasVigentes []RepartoGasto) Recibo {
	rec := Recibo{
		GastoID: g.ID,
		Tipo:    g.Tipo,
		Fecha:   g.FechaEmision,
		Total:   g.Importe,
		Lineas:  []LineaRecibo{},
	}
	if len(cuotasVigentes) == 0 {
		rec.SinReparto = true
		return rec
	}

	totalCent := int64(math.Round(g.Importe * 100))
	asignadoCent := int64(0)
	for i, c := range cuotasVigentes {
		var lineaCent int64
		if i == len(cuotasVigentes)-1 {
			// El último se lleva lo que falte para cuadrar el total al céntimo.
			lineaCent = totalCent - asignadoCent
		} else {
			lineaCent = int64(math.Round(float64(totalCent) * c.Porcentaje / PorcentajeRepartoTotal))
			asignadoCent += lineaCent
		}
		rec.Lineas = append(rec.Lineas, LineaRecibo{
			InquilinoID: c.InquilinoID,
			Porcentaje:  c.Porcentaje,
			Importe:     float64(lineaCent) / 100,
		})
	}
	return rec
}
