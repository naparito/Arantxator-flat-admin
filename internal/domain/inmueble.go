package domain

import "time"

type TipoInmueble string

const (
	TipoPiso       TipoInmueble = "piso"
	TipoCasa       TipoInmueble = "casa"
	TipoHabitacion TipoInmueble = "habitacion"
	TipoLocal      TipoInmueble = "local"
)

type EstadoInmueble string

const (
	InmuebleDisponible      EstadoInmueble = "disponible"
	InmuebleAlquilado       EstadoInmueble = "alquilado"
	InmuebleEnReforma       EstadoInmueble = "en_reforma"
	InmuebleFueraDeServicio EstadoInmueble = "fuera_de_servicio"
)

type Inmueble struct {
	ID                              int64
	Nombre                          string
	Direccion                       string
	ReferenciaCatastral             string
	CodigoPostal                    string
	Ciudad                          string
	Provincia                       string
	Tipo                            TipoInmueble
	M2Construidos                   float64
	M2Utiles                        float64
	NumHabitaciones                 int
	NumBanos                        int
	Planta                          string
	Ascensor                        bool
	Amueblado                       bool
	AnioConstruccion                int
	CertificadoEnergeticoLetra      string
	CertificadoEnergeticoCaducidad  *time.Time
	Estado                          EstadoInmueble
	CreadoEn                        time.Time
	ActualizadoEn                   time.Time
}
