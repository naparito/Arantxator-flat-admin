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

// DiasAvisoVencimiento es la ventana por defecto (en días naturales antes de
// la fecha de fin) dentro de la cual un contrato se considera "próximo a
// vencer" en lugar de "activo". Ver Contrato.EstadoDerivado.
const DiasAvisoVencimiento = 60

// DiasPlazoDepositoFianza es el plazo legal en la Comunidad de Madrid para
// depositar la fianza en la Agencia de Vivienda Social: 30 días desde la
// firma del contrato.
const DiasPlazoDepositoFianza = 30

// Contrato modela un arrendamiento sobre un Inmueble —o, si el inmueble es
// compartido, sobre una Habitacion concreta (HabitacionID no nulo)— con los
// valores por defecto de la LAU para la Comunidad de Madrid: duración mínima
// de 5 años (arrendador persona física) o 7 (persona jurídica), fianza legal
// de 1 mensualidad depositable en la Agencia de Vivienda Social, y
// actualización de renta referenciada al IRAV.
//
// El estado (activo / próximo a vencer / vencido) se deriva de las fechas al
// leer, no se guarda; solo "rescindido" se marca a mano y se persiste.
type Contrato struct {
	ID                        int64          `json:"id"`
	InmuebleID                int64          `json:"inmuebleId"`
	HabitacionID              *int64         `json:"habitacionId"`
	InquilinoIDs              []int64        `json:"inquilinoIds"`
	FechaFirma                Fecha          `json:"fechaFirma"`
	FechaInicio               Fecha          `json:"fechaInicio"`
	FechaFin                  Fecha          `json:"fechaFin"`
	ArrendadorPersonaJuridica bool           `json:"arrendadorPersonaJuridica"`
	RentaMensual              float64        `json:"rentaMensual"`
	DiaPago                   int            `json:"diaPago"`
	IndiceActualizacion       string         `json:"indiceActualizacion"`
	ProximaRevisionRenta      *Fecha         `json:"proximaRevisionRenta"`
	FianzaImporte             float64        `json:"fianzaImporte"`
	FianzaEstado              EstadoFianza   `json:"fianzaEstado"`
	FianzaFechaDeposito       *Fecha         `json:"fianzaFechaDeposito"`
	Estado                    EstadoContrato `json:"estado"`
	MotivoBaja                string         `json:"motivoBaja"`
	CreadoEn                  time.Time      `json:"creadoEn"`
	ActualizadoEn             time.Time      `json:"actualizadoEn"`

	// FechaLimiteDepositoFianza es un dato derivado (fecha de firma + 30
	// días) que se calcula al leer para el aviso de fianza de la ficha; no
	// se guarda en la tabla.
	FechaLimiteDepositoFianza Fecha `json:"fechaLimiteDepositoFianza"`
}

// soloFecha descarta la parte horaria para comparar por día natural.
func soloFecha(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// EstadoDerivado calcula el estado del contrato a partir de sus fechas y de
// una fecha de referencia (normalmente "ahora"):
//   - "rescindido" si ya se marcó así a mano (no se recalcula).
//   - "vencido" si la fecha de fin ya pasó.
//   - "proximo_a_vencer" si la fecha de fin cae dentro de la ventana de aviso.
//   - "activo" en el resto de casos.
func (c Contrato) EstadoDerivado(ref time.Time) EstadoContrato {
	if c.Estado == ContratoRescindido {
		return ContratoRescindido
	}
	hoy := soloFecha(ref)
	fin := soloFecha(time.Time(c.FechaFin))
	if fin.Before(hoy) {
		return ContratoVencido
	}
	if !fin.After(hoy.AddDate(0, 0, DiasAvisoVencimiento)) {
		return ContratoProximoAVencer
	}
	return ContratoActivo
}

// EstaVigente indica si el contrato ocupa hoy su inmueble/habitación: activo
// o próximo a vencer, pero no vencido ni rescindido. Es la condición que usa
// la regla de no solapamiento y el cálculo del % de ocupación.
func (c Contrato) EstaVigente(ref time.Time) bool {
	e := c.EstadoDerivado(ref)
	return e == ContratoActivo || e == ContratoProximoAVencer
}

// DuracionMinimaLAUAnios devuelve la duración mínima obligatoria del contrato
// según la LAU: 5 años si el arrendador es persona física, 7 si es persona
// jurídica.
func DuracionMinimaLAUAnios(arrendadorPersonaJuridica bool) int {
	if arrendadorPersonaJuridica {
		return 7
	}
	return 5
}

// FechaFinSugeridaLAU propone la fecha de fin por defecto (firma + duración
// mínima LAU) que el formulario de alta rellena y el usuario puede anular.
func FechaFinSugeridaLAU(firma time.Time, arrendadorPersonaJuridica bool) time.Time {
	return firma.AddDate(DuracionMinimaLAUAnios(arrendadorPersonaJuridica), 0, 0)
}

// FechaLimiteDepositoFianza devuelve la fecha límite para depositar la fianza
// en la Agencia de Vivienda Social: exactamente firma + 30 días.
func FechaLimiteDepositoFianza(firma time.Time) time.Time {
	return firma.AddDate(0, 0, DiasPlazoDepositoFianza)
}
