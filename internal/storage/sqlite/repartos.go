package sqlite

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/naparito/Arantxator-flat-admin/internal/domain"
)

// RepartosRepo da acceso a la tabla repartos_gasto: el reparto porcentual
// versionado de un piso compartido, por inquilino y tipo de gasto.
//
// Una "versión" del reparto de un tipo de gasto es el conjunto de filas de
// un mismo (inmueble_id, tipo_gasto, vigente_desde). Cambiar el reparto
// (CrearVersion) cierra la versión abierta anterior (le pone vigente_hasta =
// nuevo vigente_desde) e inserta las filas nuevas; nada se borra.
type RepartosRepo struct {
	db *sql.DB
}

func NewRepartosRepo(db *sql.DB) *RepartosRepo {
	return &RepartosRepo{db: db}
}

const repartoColumns = `
	id, inmueble_id, inquilino_id, tipo_gasto, porcentaje, vigente_desde, vigente_hasta, motivo, creado_en
`

func scanReparto(row interface{ Scan(...any) error }) (domain.RepartoGasto, error) {
	var (
		r      domain.RepartoGasto
		motivo sql.NullString
	)
	err := row.Scan(
		&r.ID, &r.InmuebleID, &r.InquilinoID, &r.TipoGasto, &r.Porcentaje,
		&r.VigenteDesde, &r.VigenteHasta, &motivo, &r.CreadoEn,
	)
	if err != nil {
		return domain.RepartoGasto{}, err
	}
	r.Motivo = motivo.String
	return r, nil
}

// ListByInmueble devuelve todas las filas de reparto de un inmueble (todas
// las versiones de todos los tipos de gasto), ordenadas por tipo, vigencia y
// inquilino. La agrupación en versiones y la selección de la vigente se hace
// en Go (domain.RepartoVigenteEnFecha) y en la capa HTTP.
func (r *RepartosRepo) ListByInmueble(inmuebleID int64) ([]domain.RepartoGasto, error) {
	rows, err := r.db.Query(`
		SELECT `+repartoColumns+` FROM repartos_gasto
		WHERE inmueble_id = ?
		ORDER BY tipo_gasto, vigente_desde, inquilino_id`, inmuebleID)
	if err != nil {
		return nil, fmt.Errorf("listando repartos del inmueble %d: %w", inmuebleID, err)
	}
	defer rows.Close()

	repartos := []domain.RepartoGasto{}
	for rows.Next() {
		rep, err := scanReparto(rows)
		if err != nil {
			return nil, fmt.Errorf("leyendo fila de reparto: %w", err)
		}
		repartos = append(repartos, rep)
	}
	return repartos, rows.Err()
}

// Cuota es una línea de una versión de reparto: a qué inquilino y qué %.
type Cuota struct {
	InquilinoID int64
	Porcentaje  float64
}

// CrearVersion guarda una versión nueva del reparto de un tipo de gasto:
// cierra la versión abierta anterior de ese (inmueble, tipo_gasto) poniéndole
// vigente_hasta = vigenteDesde, e inserta una fila por cuota. Todo en una
// transacción. La validación de que las cuotas sumen 100 se hace antes, en
// la capa HTTP (es una regla entre filas, no un CHECK de columna).
func (r *RepartosRepo) CrearVersion(inmuebleID int64, tipo domain.TipoGasto, vigenteDesde domain.Fecha, motivo string, cuotas []Cuota) ([]domain.RepartoGasto, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("iniciando transacción para crear reparto: %w", err)
	}
	defer tx.Rollback()

	// Cerrar las filas todavía abiertas de ese tipo de gasto que empezaron
	// antes de la nueva vigencia.
	if _, err := tx.Exec(`
		UPDATE repartos_gasto SET vigente_hasta = ?
		WHERE inmueble_id = ? AND tipo_gasto = ? AND vigente_hasta IS NULL AND vigente_desde < ?`,
		&vigenteDesde, inmuebleID, tipo, &vigenteDesde); err != nil {
		return nil, fmt.Errorf("cerrando la versión anterior del reparto: %w", err)
	}

	for _, c := range cuotas {
		if _, err := tx.Exec(`
			INSERT INTO repartos_gasto (inmueble_id, inquilino_id, tipo_gasto, porcentaje, vigente_desde, motivo)
			VALUES (?, ?, ?, ?, ?, ?)`,
			inmuebleID, c.InquilinoID, tipo, c.Porcentaje, &vigenteDesde, nullIfEmpty(motivo)); err != nil {
			return nil, fmt.Errorf("insertando cuota del inquilino %d: %w", c.InquilinoID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("confirmando creación del reparto: %w", err)
	}

	return r.VigenteEnFecha(inmuebleID, tipo, time.Time(vigenteDesde))
}

// VigenteEnFecha devuelve la versión del reparto de un tipo de gasto que
// cubre una fecha dada (una fila por inquilino), ordenada por inquilino.
func (r *RepartosRepo) VigenteEnFecha(inmuebleID int64, tipo domain.TipoGasto, fecha time.Time) ([]domain.RepartoGasto, error) {
	todos, err := r.ListByInmueble(inmuebleID)
	if err != nil {
		return nil, err
	}
	return domain.RepartoVigenteEnFecha(todos, tipo, fecha), nil
}
