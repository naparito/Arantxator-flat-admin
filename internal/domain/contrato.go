package domain

import "time"

type EstadoContrato string

const (
	ContratoActivo         EstadoContrato = "activo"
	ContratoProximoAVencer EstadoContrato = "proximo_a_vencer"
	ContratoVencido        EstadoContrato = "vencido"
	ContratoRescindido     EstadoContrato = "rescindido"
)

type EstadoFianza string

const (
	FianzaPendiente    EstadoFianza = "pendiente"
	FianzaDepositada   EstadoFianza = "depositada"
	FianzaEnDevolucion EstadoFianza = "en_devolucion"
	FianzaDevuelta     EstadoFianza = "devuelta"
)

// Contrato modela un arrendamiento sobre un Inmueble con los valores por
// defecto de la LAU para la Comunidad de Madrid: duración mínima de 5 años
// (arrendador persona física) o 7 (persona jurídica), fianza legal de 1
// mensualidad depositable en la Agencia de Vivienda Social, y actualización
// de renta referenciada al IRAV.
type Contrato struct {
	ID                        int64
	InmuebleID                int64
	InquilinoIDs              []int64
	FechaFirma                time.Time
	FechaInicio               time.Time
	FechaFin                  time.Time
	ArrendadorPersonaJuridica bool
	RentaMensual              float64
	DiaPago                   int
	IndiceActualizacion       string
	ProximaRevisionRenta      *time.Time
	FianzaImporte             float64
	FianzaEstado              EstadoFianza
	FianzaFechaDeposito       *time.Time
	Estado                    EstadoContrato
	MotivoBaja                string
	CreadoEn                  time.Time
	ActualizadoEn             time.Time
}
