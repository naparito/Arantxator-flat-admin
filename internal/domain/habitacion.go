package domain

import "time"

// Habitacion es la unidad de control físico de un inmueble compartido
// (ver Inmueble.Compartido): cada habitación se da de alta por separado,
// con sus propias características. La asignación de ocupante es de
// presentación (quién vive hoy ahí), no la relación contractual — esa
// sigue siendo Contrato-Inquilino.
type Habitacion struct {
	ID            int64     `json:"id"`
	InmuebleID    int64     `json:"inmuebleId"`
	Nombre        string    `json:"nombre"`
	M2            float64   `json:"m2"`
	TieneBano     bool      `json:"tieneBano"`
	Amueblada     bool      `json:"amueblada"`
	Notas         string    `json:"notas"`
	InquilinoID   *int64    `json:"inquilinoId"`
	CreadoEn      time.Time `json:"creadoEn"`
	ActualizadoEn time.Time `json:"actualizadoEn"`
}
