package domain

import "time"

// Inquilino guarda los datos personales, de contacto y de pago de un
// inquilino. El IBAN se persiste completo; el enmascarado (ej.
// "ES91 •••• •••• 1234") es solo de presentación en la GUI.
type Inquilino struct {
	ID                         int64      `json:"id"`
	NombreCompleto             string     `json:"nombreCompleto"`
	DocumentoIdentidad         string     `json:"documentoIdentidad"`
	FechaNacimiento            *time.Time `json:"fechaNacimiento"`
	Telefono                   string     `json:"telefono"`
	Email                      string     `json:"email"`
	Nacionalidad               string     `json:"nacionalidad"`
	ContactoEmergenciaNombre   string     `json:"contactoEmergenciaNombre"`
	ContactoEmergenciaTelefono string     `json:"contactoEmergenciaTelefono"`
	IBAN                       string     `json:"iban"`
	CreadoEn                   time.Time  `json:"creadoEn"`
	ActualizadoEn              time.Time  `json:"actualizadoEn"`
}
