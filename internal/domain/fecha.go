package domain

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

// fechaLayout es el único formato de fecha (sin hora) que entra y sale por
// la API y por la base de datos: "2006-01-02", el mismo que produce y
// espera un <input type="date"> de la GUI.
const fechaLayout = "2006-01-02"

// Fecha envuelve time.Time para los campos de solo fecha (fecha de
// nacimiento, caducidad de certificados, fechas de contrato del Hito 3...).
// El (Un)marshalJSON por defecto de time.Time exige RFC3339 completo (con
// hora y zona), que es justo lo que NO manda un <input type="date"> — sin
// este tipo, el JSON que llega de la GUI para estos campos no deserializa y
// la API responde "cuerpo de la petición inválido" con cualquier fecha.
type Fecha time.Time

func (f Fecha) MarshalJSON() ([]byte, error) {
	return []byte(`"` + time.Time(f).Format(fechaLayout) + `"`), nil
}

func (f *Fecha) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	if s == "" || s == "null" {
		return nil
	}
	t, err := time.Parse(fechaLayout, s)
	if err != nil {
		return fmt.Errorf("fecha %q inválida, se espera AAAA-MM-DD: %w", s, err)
	}
	*f = Fecha(t)
	return nil
}

// Scan implementa sql.Scanner: el driver sqlite puede devolver la columna
// DATE como time.Time o como texto, según cómo se guardó.
func (f *Fecha) Scan(value any) error {
	if value == nil {
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		*f = Fecha(v)
		return nil
	case string:
		if t, err := time.Parse(fechaLayout, v); err == nil {
			*f = Fecha(t)
			return nil
		}
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return fmt.Errorf("escaneando fecha %q: %w", v, err)
		}
		*f = Fecha(t)
		return nil
	default:
		return fmt.Errorf("tipo inesperado para Fecha: %T", value)
	}
}

// Value implementa driver.Valuer para persistir la fecha en sqlite.
func (f *Fecha) Value() (driver.Value, error) {
	if f == nil {
		return nil, nil
	}
	return time.Time(*f), nil
}
