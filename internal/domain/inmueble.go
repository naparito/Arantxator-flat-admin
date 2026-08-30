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

	// Ocupacion es un dato derivado en lectura (no una columna): solo se
	// rellena para inmuebles compartidos, con el % de habitaciones que
	// tienen un contrato vigente. En un inmueble no compartido es nil y el
	// estado operativo binario (Estado) ya representa la ocupación.
	Ocupacion *OcupacionInmueble `json:"ocupacion,omitempty"`
}

// OcupacionInmueble resume la ocupación de un inmueble compartido:
// habitaciones con al menos un contrato vigente sobre el total, y su
// porcentaje redondeado (ej. 2/3 -> 67%). Se calcula al leer, cruzando
// contratos y habitaciones; el frontend no tiene que recomponerlo.
type OcupacionInmueble struct {
	HabitacionesTotales  int `json:"habitacionesTotales"`
	HabitacionesOcupadas int `json:"habitacionesOcupadas"`
	Porcentaje           int `json:"porcentaje"`
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
