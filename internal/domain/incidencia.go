package domain

import "time"

type PrioridadIncidencia string

const (
	PrioridadBaja    PrioridadIncidencia = "baja"
	PrioridadMedia   PrioridadIncidencia = "media"
	PrioridadAlta    PrioridadIncidencia = "alta"
	PrioridadUrgente PrioridadIncidencia = "urgente"
)

type EstadoIncidencia string

const (
	IncidenciaAbierta            EstadoIncidencia = "abierta"
	IncidenciaEnProceso          EstadoIncidencia = "en_proceso"
	IncidenciaEsperandoProveedor EstadoIncidencia = "esperando_proveedor"
	IncidenciaResuelta           EstadoIncidencia = "resuelta"
	IncidenciaCerrada            EstadoIncidencia = "cerrada"
)

type CosteACargoDe string

const (
	CostePropietario CosteACargoDe = "propietario"
	CosteInquilino   CosteACargoDe = "inquilino"
)

type Incidencia struct {
	ID                int64
	InmuebleID        int64
	Titulo            string
	Descripcion       string
	Categoria         string
	Prioridad         PrioridadIncidencia
	Origen            string
	Estado            EstadoIncidencia
	ProveedorNombre   string
	ProveedorContacto string
	Coste             float64
	CosteACargoDe     CosteACargoDe
	FechaApertura     time.Time
	FechaCierre       *time.Time
	CreadoEn          time.Time
	ActualizadoEn     time.Time
}
