package domain

// ResumenDashboard es la foto agregada de toda la cartera que sirve
// GET /api/dashboard/resumen (§8 del diseño técnico-funcional). Todos sus
// números se recalculan al leer cruzando los repos reales (contratos,
// gastos, incidencias, cobros, inmuebles) — nunca sobre una cifra cacheada.
// El frontend los pinta tal cual, sin volver a cruzar datos por su cuenta.
type ResumenDashboard struct {
	Periodo               string                  `json:"periodo"` // "AAAA-MM" de la rentabilidad agregada
	Ocupacion             OcupacionCartera        `json:"ocupacion"`
	ContratosPorVencer    int                     `json:"contratosPorVencer"`
	GastosPendientes      GastosPendientesResumen `json:"gastosPendientes"`
	IncidenciasAbiertas   int                     `json:"incidenciasAbiertas"`
	Rentabilidad          RentabilidadResumen     `json:"rentabilidad"`
	NotificacionesSinLeer int                     `json:"notificacionesSinLeer"`
}

// OcupacionCartera resume la ocupación del conjunto de inmuebles. Un inmueble
// cuenta como "ocupado" si no es compartido y su estado es "alquilado", o si
// es compartido y tiene al menos una habitación con contrato vigente. El
// desglose por habitaciones suma solo las de los inmuebles compartidos.
type OcupacionCartera struct {
	InmueblesTotales     int `json:"inmueblesTotales"`
	InmueblesOcupados    int `json:"inmueblesOcupados"`
	HabitacionesTotales  int `json:"habitacionesTotales"`
	HabitacionesOcupadas int `json:"habitacionesOcupadas"`
	Porcentaje           int `json:"porcentaje"` // inmuebles ocupados / totales, redondeado
}

// GastosPendientesResumen agrega las facturas cuyo estado de pago derivado no
// es "pagado" (pendientes + vencidas): cuántas son y cuánto suman.
type GastosPendientesResumen struct {
	Cantidad int     `json:"cantidad"`
	Importe  float64 `json:"importe"`
}

// RentabilidadResumen es la rentabilidad neta agregada de toda la cartera en
// un mes: ingresos por renta cobrada − gastos con fecha de emisión en el mes.
// Es la misma cuenta que domain.Rentabilidad por inmueble, sumada.
type RentabilidadResumen struct {
	Periodo  string  `json:"periodo"`
	Ingresos float64 `json:"ingresos"`
	Gastos   float64 `json:"gastos"`
	Neto     float64 `json:"neto"`
}
