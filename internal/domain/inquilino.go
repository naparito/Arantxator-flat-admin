package domain

import "time"

type Inquilino struct {
	ID                         int64
	NombreCompleto             string
	DocumentoIdentidad         string
	FechaNacimiento            *time.Time
	Telefono                   string
	Email                      string
	Nacionalidad               string
	ContactoEmergenciaNombre   string
	ContactoEmergenciaTelefono string
	IBAN                       string
	CreadoEn                   time.Time
	ActualizadoEn              time.Time
}
