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
	ID                             int64          `json:"id"`
	Nombre                         string         `json:"nombre"`
	Direccion                      string         `json:"direccion"`
	ReferenciaCatastral            string         `json:"referenciaCatastral"`
	CodigoPostal                   string         `json:"codigoPostal"`
	Ciudad                         string         `json:"ciudad"`
	Provincia                      string         `json:"provincia"`
	Tipo                           TipoInmueble   `json:"tipo"`
	M2Construidos                  float64        `json:"m2Construidos"`
	M2Utiles                       float64        `json:"m2Utiles"`
	NumHabitaciones                int            `json:"numHabitaciones"`
	NumBanos                       int            `json:"numBanos"`
	Planta                         string         `json:"planta"`
	Ascensor                       bool           `json:"ascensor"`
	Amueblado                      bool           `json:"amueblado"`
	AnioConstruccion               int            `json:"anioConstruccion"`
	CertificadoEnergeticoLetra     string         `json:"certificadoEnergeticoLetra"`
	CertificadoEnergeticoCaducidad *Fecha         `json:"certificadoEnergeticoCaducidad"`
	Estado                         EstadoInmueble `json:"estado"`
	// Compartido indica que el inmueble se alquila por habitaciones en
	// lugar de a un único arrendatario; activa el submódulo de
	// habitaciones (ver domain.Habitacion).
	Compartido    bool        `json:"compartido"`
	Suministros   Suministros `json:"suministros"`
	CreadoEn      time.Time   `json:"creadoEn"`
	ActualizadoEn time.Time   `json:"actualizadoEn"`
}

// Suministro guarda los datos de contratación de una compañía suministradora
// (luz, agua, gas o internet) para un inmueble.
type Suministro struct {
	Compania       string `json:"compania"`
	NumeroContrato string `json:"numeroContrato"`
	Titular        string `json:"titular"`
}

type Suministros struct {
	Luz      Suministro `json:"luz"`
	Agua     Suministro `json:"agua"`
	Gas      Suministro `json:"gas"`
	Internet Suministro `json:"internet"`
}
