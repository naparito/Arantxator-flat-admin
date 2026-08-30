package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/naparito/Arantxator-flat-admin/internal/domain"
)

// CobrosRepo da acceso a la tabla cobros_renta: el registro de los cobros
// mensuales de renta de un inmueble (§7.3), que alimenta el término
// "ingresos" de la rentabilidad neta.
type CobrosRepo struct {
	db *sql.DB
}

func NewCobrosRepo(db *sql.DB) *CobrosRepo {
	return &CobrosRepo{db: db}
}

const cobroColumns = `
	id, inmueble_id, contrato_id, periodo, importe, fecha_cobro, metodo_pago, notas, creado_en, actualizado_en
`

func scanCobro(row interface{ Scan(...any) error }) (domain.CobroRenta, error) {
	var c domain.CobroRenta
	err := row.Scan(
		&c.ID, &c.InmuebleID, &c.ContratoID, &c.Periodo, &c.Importe, &c.FechaCobro,
		&c.MetodoPago, &c.Notas, &c.CreadoEn, &c.ActualizadoEn,
	)
	return c, err
}

func (r *CobrosRepo) Create(c domain.CobroRenta) (domain.CobroRenta, error) {
	res, err := r.db.Exec(`
		INSERT INTO cobros_renta (inmueble_id, contrato_id, periodo, importe, fecha_cobro, metodo_pago, notas)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		c.InmuebleID, c.ContratoID, &c.Periodo, c.Importe, c.FechaCobro, c.MetodoPago, c.Notas,
	)
	if err != nil {
		return domain.CobroRenta{}, fmt.Errorf("insertando cobro de renta: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return domain.CobroRenta{}, fmt.Errorf("obteniendo id del cobro creado: %w", err)
	}
	return r.Get(id)
}

func (r *CobrosRepo) Get(id int64) (domain.CobroRenta, error) {
	row := r.db.QueryRow(`SELECT `+cobroColumns+` FROM cobros_renta WHERE id = ?`, id)
	c, err := scanCobro(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.CobroRenta{}, ErrNotFound
	}
	if err != nil {
		return domain.CobroRenta{}, fmt.Errorf("leyendo cobro %d: %w", id, err)
	}
	return c, nil
}

// ListByInmueble devuelve los cobros de un inmueble, los de periodo más
// reciente primero.
func (r *CobrosRepo) ListByInmueble(inmuebleID int64) ([]domain.CobroRenta, error) {
	rows, err := r.db.Query(`
		SELECT `+cobroColumns+` FROM cobros_renta
		WHERE inmueble_id = ?
		ORDER BY periodo DESC, id DESC`, inmuebleID)
	if err != nil {
		return nil, fmt.Errorf("listando cobros del inmueble %d: %w", inmuebleID, err)
	}
	defer rows.Close()

	cobros := []domain.CobroRenta{}
	for rows.Next() {
		c, err := scanCobro(rows)
		if err != nil {
			return nil, fmt.Errorf("leyendo fila de cobro: %w", err)
		}
		cobros = append(cobros, c)
	}
	return cobros, rows.Err()
}

func (r *CobrosRepo) Update(id int64, c domain.CobroRenta) (domain.CobroRenta, error) {
	res, err := r.db.Exec(`
		UPDATE cobros_renta SET
			contrato_id = ?, periodo = ?, importe = ?, fecha_cobro = ?, metodo_pago = ?, notas = ?,
			actualizado_en = CURRENT_TIMESTAMP
		WHERE id = ?`,
		c.ContratoID, &c.Periodo, c.Importe, c.FechaCobro, c.MetodoPago, c.Notas, id,
	)
	if err != nil {
		return domain.CobroRenta{}, fmt.Errorf("actualizando cobro %d: %w", id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return domain.CobroRenta{}, fmt.Errorf("comprobando actualización de cobro %d: %w", id, err)
	}
	if n == 0 {
		return domain.CobroRenta{}, ErrNotFound
	}
	return r.Get(id)
}

// SumaEnPeriodo suma los importes de los cobros de renta de un inmueble cuyo
// periodo cae en [desde, hasta) — el término "ingresos" de la rentabilidad.
func (r *CobrosRepo) SumaEnPeriodo(inmuebleID int64, desde, hasta time.Time) (float64, error) {
	cobros, err := r.ListByInmueble(inmuebleID)
	if err != nil {
		return 0, err
	}
	d, h := soloDia(desde), soloDia(hasta)
	total := 0.0
	for _, c := range cobros {
		p := soloDia(time.Time(c.Periodo))
		if !p.Before(d) && p.Before(h) {
			total += c.Importe
		}
	}
	return total, nil
}
