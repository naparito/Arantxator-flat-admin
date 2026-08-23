package domain

import "time"

type EntidadTipo string

const (
	EntidadInmueble   EntidadTipo = "inmueble"
	EntidadInquilino  EntidadTipo = "inquilino"
	EntidadContrato   EntidadTipo = "contrato"
	EntidadGasto      EntidadTipo = "gasto"
	EntidadIncidencia EntidadTipo = "incidencia"
)

// Documento guarda cualquier adjunto (contrato firmado, factura, foto) como
// BLOB dentro de la propia base de datos: la copia de seguridad de toda la
// aplicación es siempre un único fichero .db.
type Documento struct {
	ID            int64
	EntidadTipo   EntidadTipo
	EntidadID     int64
	NombreArchivo string
	TipoMIME      string
	Contenido     []byte
	TamanoBytes   int64
	SubidoEn      time.Time
}
