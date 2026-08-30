package domain

import "time"

// CobroRenta registra un cobro mensual de renta de un inmueble. El §7.3 del
// diseño técnico-funcional dice literalmente que "el cobro mensual de renta
// se registra junto a los gastos" para poder calcular la rentabilidad neta
// (ingresos − gastos): por eso es una entidad propia con su importe y su
// fecha real de cobro, no un dato derivado de la renta teórica de los
// contratos (que no distingue "renta debida" de "renta efectivamente
// cobrada", que es lo que mide la rentabilidad neta).
//
// Periodo es el primer día del mes al que corresponde la renta (una columna
// DATE); FechaCobro es el día real en que entró el dinero. Ambas son
// domain.Fecha (solo fecha), no *time.Time.
type CobroRenta struct {
	ID            int64     `json:"id"`
	InmuebleID    int64     `json:"inmuebleId"`
	ContratoID    *int64    `json:"contratoId"`
	Periodo       Fecha     `json:"periodo"`
	Importe       float64   `json:"importe"`
	FechaCobro    *Fecha    `json:"fechaCobro"`
	MetodoPago    string    `json:"metodoPago"`
	Notas         string    `json:"notas"`
	CreadoEn      time.Time `json:"creadoEn"`
	ActualizadoEn time.Time `json:"actualizadoEn"`
}

// Rentabilidad resume, para un inmueble y un periodo (un mes natural), los
// ingresos por renta cobrada, los gastos con fecha de emisión dentro del
// periodo y su diferencia. Se calcula al leer sobre los datos reales; el
// frontend no tiene que cruzar gastos y cobros por su cuenta.
type Rentabilidad struct {
	InmuebleID int64   `json:"inmuebleId"`
	Periodo    string  `json:"periodo"` // "AAAA-MM"
	Ingresos   float64 `json:"ingresos"`
	Gastos     float64 `json:"gastos"`
	Neto       float64 `json:"neto"`
}
