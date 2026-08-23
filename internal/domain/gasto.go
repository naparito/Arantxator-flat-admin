package domain

import "time"

type EstadoPago string

const (
	PagoPendiente EstadoPago = "pendiente"
	PagoPagado    EstadoPago = "pagado"
	PagoVencido   EstadoPago = "vencido"
)

type Gasto struct {
	ID               int64
	InmuebleID       int64
	Tipo             string
	Periodicidad     string
	Importe          float64
	FechaEmision     time.Time
	FechaVencimiento *time.Time
	Proveedor        string
	EstadoPago       EstadoPago
	FechaPago        *time.Time
	MetodoPago       string
	CreadoEn         time.Time
	ActualizadoEn    time.Time
}

// RepartoGasto vive como entidad propia, independiente del Inmueble, para
// poder versionar en el tiempo qué porcentaje corresponde a cada inquilino
// por tipo de gasto (cambia cuando entra o sale alguien, o cuando se
// renegocia el reparto de un gasto concreto).
type RepartoGasto struct {
	ID           int64
	InmuebleID   int64
	InquilinoID  int64
	TipoGasto    string
	Porcentaje   float64
	VigenteDesde time.Time
	VigenteHasta *time.Time
	CreadoEn     time.Time
}
