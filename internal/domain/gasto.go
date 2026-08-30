package domain

import "time"

type EstadoPago string

const (
	PagoPendiente EstadoPago = "pendiente"
	PagoPagado    EstadoPago = "pagado"
	PagoVencido   EstadoPago = "vencido"
)

// TipoGasto enumera los tipos de factura por inmueble (§7.1 del diseño
// técnico-funcional). La columna gastos.tipo no tiene CHECK en el esquema
// inicial, así que la validación es a nivel de aplicación (igual que el tipo
// de gasto de un reparto).
type TipoGasto string

const (
	GastoAgua          TipoGasto = "agua"
	GastoLuz           TipoGasto = "luz"
	GastoGas           TipoGasto = "gas"
	GastoInternet      TipoGasto = "internet"
	GastoComunidad     TipoGasto = "comunidad"
	GastoIBI           TipoGasto = "ibi"
	GastoSeguro        TipoGasto = "seguro"
	GastoMantenimiento TipoGasto = "mantenimiento"
	GastoBasuras       TipoGasto = "basuras"
	GastoGestoria      TipoGasto = "gestoria"
	GastoOtros         TipoGasto = "otros"
)

// TiposGasto lista los valores válidos, en el mismo orden que la tabla del
// §7.1. Sirve para validar la entrada de la API y para las columnas de la
// matriz de reparto de la GUI.
var TiposGasto = []TipoGasto{
	GastoAgua, GastoLuz, GastoGas, GastoInternet, GastoComunidad, GastoIBI,
	GastoSeguro, GastoMantenimiento, GastoBasuras, GastoGestoria, GastoOtros,
}

func (t TipoGasto) Valida() bool {
	for _, v := range TiposGasto {
		if v == t {
			return true
		}
	}
	return false
}

// Periodicidad es cada cuánto se emite un gasto recurrente (§7.1).
// Es opcional: un gasto puntual puede dejarla vacía.
type Periodicidad string

const (
	PeriodicidadMensual    Periodicidad = "mensual"
	PeriodicidadBimestral  Periodicidad = "bimestral"
	PeriodicidadTrimestral Periodicidad = "trimestral"
	PeriodicidadAnual      Periodicidad = "anual"
)

var Periodicidades = []Periodicidad{
	PeriodicidadMensual, PeriodicidadBimestral, PeriodicidadTrimestral, PeriodicidadAnual,
}

func (p Periodicidad) Valida() bool {
	for _, v := range Periodicidades {
		if v == p {
			return true
		}
	}
	return false
}

// Gasto modela una factura por inmueble (§7.1): tipo, periodicidad, importe,
// fechas de emisión/vencimiento/pago, proveedor y estado de pago, con PDF
// adjunto vía el endpoint genérico de documentos (entidad_tipo = 'gasto').
//
// fecha_emision, fecha_vencimiento y fecha_pago son columnas DATE (solo
// fecha), por eso son domain.Fecha y no *time.Time: un *time.Time desnudo
// rompe con un 400 genérico en cuanto la GUI manda una fecha desde un
// <input type="date"> (mismo bug corregido en los hitos 2, 3 y 4).
// creado_en/actualizado_en sí son DATETIME → time.Time.
//
// El estado de pago se guarda solo como "pendiente" o "pagado"; "vencido"
// es un valor derivado al leer (pendiente + fecha de vencimiento pasada),
// igual que el estado de un contrato. Ver EstadoPagoDerivado.
type Gasto struct {
	ID               int64        `json:"id"`
	InmuebleID       int64        `json:"inmuebleId"`
	Tipo             TipoGasto    `json:"tipo"`
	Periodicidad     Periodicidad `json:"periodicidad"`
	Importe          float64      `json:"importe"`
	FechaEmision     Fecha        `json:"fechaEmision"`
	FechaVencimiento *Fecha       `json:"fechaVencimiento"`
	Proveedor        string       `json:"proveedor"`
	EstadoPago       EstadoPago   `json:"estadoPago"`
	FechaPago        *Fecha       `json:"fechaPago"`
	MetodoPago       string       `json:"metodoPago"`
	CreadoEn         time.Time    `json:"creadoEn"`
	ActualizadoEn    time.Time    `json:"actualizadoEn"`
}

// EstadoPagoDerivado calcula el estado que se muestra en la GUI a partir del
// guardado y de una fecha de referencia ("ahora"):
//   - "pagado" si ya se marcó así (no se recalcula).
//   - "vencido" si sigue pendiente y la fecha de vencimiento ya pasó.
//   - "pendiente" en el resto de casos.
func (g Gasto) EstadoPagoDerivado(ref time.Time) EstadoPago {
	if g.EstadoPago == PagoPagado {
		return PagoPagado
	}
	if g.FechaVencimiento == nil {
		return PagoPendiente
	}
	hoy := soloFecha(ref)
	venc := soloFecha(time.Time(*g.FechaVencimiento))
	if venc.Before(hoy) {
		return PagoVencido
	}
	return PagoPendiente
}

// RepartoGasto vive como entidad propia, independiente del Inmueble, para
// poder versionar en el tiempo qué porcentaje corresponde a cada inquilino
// por tipo de gasto (cambia cuando entra o sale alguien, o cuando se
// renegocia el reparto de un gasto concreto).
//
// Una "versión" del reparto es el conjunto de filas de un mismo
// (inmueble_id, tipo_gasto, vigente_desde): cada inquilino activo con su %.
// Cambiar el reparto crea una versión nueva con su vigente_desde y cierra la
// anterior poniéndole vigente_hasta = ese vigente_desde (intervalos
// semiabiertos [desde, hasta)); la versión anterior no se borra.
//
// vigente_desde y vigente_hasta son DATE → domain.Fecha. Motivo es la nota
// libre que explica el cambio ("entrada de Pablo Navarro"), que la GUI
// muestra junto a la matriz.
type RepartoGasto struct {
	ID           int64     `json:"id"`
	InmuebleID   int64     `json:"inmuebleId"`
	InquilinoID  int64     `json:"inquilinoId"`
	TipoGasto    TipoGasto `json:"tipoGasto"`
	Porcentaje   float64   `json:"porcentaje"`
	VigenteDesde Fecha     `json:"vigenteDesde"`
	VigenteHasta *Fecha    `json:"vigenteHasta"`
	Motivo       string    `json:"motivo"`
	CreadoEn     time.Time `json:"creadoEn"`
}
