package domain

import "time"

type PrioridadIncidencia string

const (
	PrioridadBaja    PrioridadIncidencia = "baja"
	PrioridadMedia   PrioridadIncidencia = "media"
	PrioridadAlta    PrioridadIncidencia = "alta"
	PrioridadUrgente PrioridadIncidencia = "urgente"
)

// PrioridadesIncidencia enumera los valores válidos (mismo orden que el
// CHECK de la tabla). Sirve para validar la entrada de la API.
var PrioridadesIncidencia = []PrioridadIncidencia{
	PrioridadBaja, PrioridadMedia, PrioridadAlta, PrioridadUrgente,
}

func (p PrioridadIncidencia) Valida() bool {
	for _, v := range PrioridadesIncidencia {
		if v == p {
			return true
		}
	}
	return false
}

type EstadoIncidencia string

const (
	IncidenciaAbierta            EstadoIncidencia = "abierta"
	IncidenciaEnProceso          EstadoIncidencia = "en_proceso"
	IncidenciaEsperandoProveedor EstadoIncidencia = "esperando_proveedor"
	IncidenciaResuelta           EstadoIncidencia = "resuelta"
	IncidenciaCerrada            EstadoIncidencia = "cerrada"
)

// FlujoIncidencia es el recorrido de estados en orden: cada incidencia
// nace "abierta" y avanza un paso cada vez hasta "cerrada". Ver
// TransicionEstadoIncidenciaValida y §4.3 del diseño técnico-funcional.
var FlujoIncidencia = []EstadoIncidencia{
	IncidenciaAbierta,
	IncidenciaEnProceso,
	IncidenciaEsperandoProveedor,
	IncidenciaResuelta,
	IncidenciaCerrada,
}

func (e EstadoIncidencia) Valido() bool {
	return indiceEstadoIncidencia(e) >= 0
}

func indiceEstadoIncidencia(e EstadoIncidencia) int {
	for i, v := range FlujoIncidencia {
		if v == e {
			return i
		}
	}
	return -1
}

// TransicionEstadoIncidenciaValida indica si el usuario puede pasar una
// incidencia de `desde` a `hasta` respetando el flujo
// abierta → en proceso → esperando proveedor → resuelta → cerrada:
//   - avanzar exactamente un paso en el flujo, o
//   - reabrir: una incidencia "resuelta" o "cerrada" puede volver a
//     "en proceso" si el problema reaparece.
//
// No se permite saltarse pasos ni "cambiar" al mismo estado.
func TransicionEstadoIncidenciaValida(desde, hasta EstadoIncidencia) bool {
	i, j := indiceEstadoIncidencia(desde), indiceEstadoIncidencia(hasta)
	if i < 0 || j < 0 || i == j {
		return false
	}
	if j == i+1 {
		return true
	}
	if (desde == IncidenciaResuelta || desde == IncidenciaCerrada) && hasta == IncidenciaEnProceso {
		return true
	}
	return false
}

// EsFinal indica si el estado da la incidencia por terminada (deja de
// contar en el badge de "incidencias abiertas" del tab).
func (e EstadoIncidencia) EsFinal() bool {
	return e == IncidenciaCerrada
}

type CosteACargoDe string

const (
	CostePropietario CosteACargoDe = "propietario"
	CosteInquilino   CosteACargoDe = "inquilino"
)

func (c CosteACargoDe) Valido() bool {
	return c == CostePropietario || c == CosteInquilino
}

type OrigenIncidencia string

const (
	OrigenInquilino   OrigenIncidencia = "inquilino"
	OrigenPropietario OrigenIncidencia = "propietario"
)

func (o OrigenIncidencia) Valido() bool {
	return o == OrigenInquilino || o == OrigenPropietario
}

// Incidencia es un parte de mantenimiento sobre un inmueble (§4.3): nace
// "abierta" y recorre el flujo de estados a mano, dejando cada cambio
// registrado con su fecha en IncidenciaEvento. FechaApertura y FechaCierre
// son marcas de tiempo (DATETIME), por eso son time.Time y no domain.Fecha.
type Incidencia struct {
	ID                int64               `json:"id"`
	InmuebleID        int64               `json:"inmuebleId"`
	Titulo            string              `json:"titulo"`
	Descripcion       string              `json:"descripcion"`
	Categoria         string              `json:"categoria"`
	Prioridad         PrioridadIncidencia `json:"prioridad"`
	Origen            OrigenIncidencia    `json:"origen"`
	Estado            EstadoIncidencia    `json:"estado"`
	ProveedorNombre   string              `json:"proveedorNombre"`
	ProveedorContacto string              `json:"proveedorContacto"`
	Coste             float64             `json:"coste"`
	CosteACargoDe     CosteACargoDe       `json:"costeACargoDe"`
	FechaApertura     time.Time           `json:"fechaApertura"`
	FechaCierre       *time.Time          `json:"fechaCierre"`
	CreadoEn          time.Time           `json:"creadoEn"`
	ActualizadoEn     time.Time           `json:"actualizadoEn"`

	// Eventos es el historial fechado de la incidencia (alta, cada cambio
	// de estado y los comentarios de seguimiento), más reciente al final.
	// Se rellena al leer; no es una columna.
	Eventos []IncidenciaEvento `json:"eventos"`

	// Comentario es un campo de solo-entrada: si el PUT lo trae con texto,
	// se añade como un evento de tipo "comentario". Nunca se persiste en la
	// fila de la incidencia (de ahí el omitempty y que Get lo devuelva vacío).
	Comentario string `json:"comentario,omitempty"`
}

type TipoEventoIncidencia string

const (
	EventoAlta         TipoEventoIncidencia = "alta"
	EventoCambioEstado TipoEventoIncidencia = "cambio_estado"
	EventoComentario   TipoEventoIncidencia = "comentario"
)

// IncidenciaEvento es una entrada del historial de una incidencia. Para
// 'cambio_estado' van EstadoAnterior/EstadoNuevo; para 'comentario', el
// texto en Comentario; 'alta' solo lleva EstadoNuevo ("abierta").
type IncidenciaEvento struct {
	ID             int64                `json:"id"`
	IncidenciaID   int64                `json:"incidenciaId"`
	Tipo           TipoEventoIncidencia `json:"tipo"`
	EstadoAnterior EstadoIncidencia     `json:"estadoAnterior,omitempty"`
	EstadoNuevo    EstadoIncidencia     `json:"estadoNuevo,omitempty"`
	Comentario     string               `json:"comentario,omitempty"`
	CreadoEn       time.Time            `json:"creadoEn"`
}
