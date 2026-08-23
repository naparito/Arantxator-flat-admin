package sqlite

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/naparito/Arantxator-flat-admin/internal/domain"
)

// HabitacionesRepo da acceso a la tabla habitaciones (submódulo de
// inmuebles compartidos).
type HabitacionesRepo struct {
	db *sql.DB
}

func NewHabitacionesRepo(db *sql.DB) *HabitacionesRepo {
	return &HabitacionesRepo{db: db}
}

const habitacionColumns = `id, inmueble_id, nombre, m2, tiene_bano, amueblada, notas, inquilino_id, creado_en, actualizado_en`

func scanHabitacion(row interface{ Scan(...any) error }) (domain.Habitacion, error) {
	var h domain.Habitacion
	err := row.Scan(&h.ID, &h.InmuebleID, &h.Nombre, &h.M2, &h.TieneBano, &h.Amueblada, &h.Notas, &h.InquilinoID, &h.CreadoEn, &h.ActualizadoEn)
	return h, err
}

// Create inserta una habitación nueva para un inmueble y devuelve el
// registro tal como queda persistido.
func (r *HabitacionesRepo) Create(h domain.Habitacion) (domain.Habitacion, error) {
	res, err := r.db.Exec(`
		INSERT INTO habitaciones (inmueble_id, nombre, m2, tiene_bano, amueblada, notas, inquilino_id)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		h.InmuebleID, h.Nombre, h.M2, h.TieneBano, h.Amueblada, h.Notas, h.InquilinoID,
	)
	if err != nil {
		return domain.Habitacion{}, fmt.Errorf("insertando habitación: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return domain.Habitacion{}, fmt.Errorf("obteniendo id de la habitación creada: %w", err)
	}
	return r.Get(id)
}

// Get devuelve una habitación por id, o ErrNotFound si no existe.
func (r *HabitacionesRepo) Get(id int64) (domain.Habitacion, error) {
	row := r.db.QueryRow(`SELECT `+habitacionColumns+` FROM habitaciones WHERE id = ?`, id)
	h, err := scanHabitacion(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Habitacion{}, ErrNotFound
	}
	if err != nil {
		return domain.Habitacion{}, fmt.Errorf("leyendo habitación %d: %w", id, err)
	}
	return h, nil
}

// ListByInmueble devuelve las habitaciones de un inmueble, ordenadas por
// nombre.
func (r *HabitacionesRepo) ListByInmueble(inmuebleID int64) ([]domain.Habitacion, error) {
	rows, err := r.db.Query(`SELECT `+habitacionColumns+` FROM habitaciones WHERE inmueble_id = ? ORDER BY nombre`, inmuebleID)
	if err != nil {
		return nil, fmt.Errorf("listando habitaciones del inmueble %d: %w", inmuebleID, err)
	}
	defer rows.Close()

	habitaciones := []domain.Habitacion{}
	for rows.Next() {
		h, err := scanHabitacion(rows)
		if err != nil {
			return nil, fmt.Errorf("leyendo fila de habitación: %w", err)
		}
		habitaciones = append(habitaciones, h)
	}
	return habitaciones, rows.Err()
}

// Update sobrescribe las características editables de una habitación
// (no toca el ocupante; ver AsignarOcupante). Devuelve ErrNotFound si el
// id no existe.
func (r *HabitacionesRepo) Update(id int64, h domain.Habitacion) (domain.Habitacion, error) {
	res, err := r.db.Exec(`
		UPDATE habitaciones SET nombre = ?, m2 = ?, tiene_bano = ?, amueblada = ?, notas = ?, actualizado_en = CURRENT_TIMESTAMP
		WHERE id = ?`,
		h.Nombre, h.M2, h.TieneBano, h.Amueblada, h.Notas, id,
	)
	if err != nil {
		return domain.Habitacion{}, fmt.Errorf("actualizando habitación %d: %w", id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return domain.Habitacion{}, fmt.Errorf("comprobando actualización de habitación %d: %w", id, err)
	}
	if n == 0 {
		return domain.Habitacion{}, ErrNotFound
	}
	return r.Get(id)
}

// AsignarOcupante asigna (o, con inquilinoID nulo, quita) el inquilino
// ocupante de una habitación.
func (r *HabitacionesRepo) AsignarOcupante(id int64, inquilinoID *int64) (domain.Habitacion, error) {
	res, err := r.db.Exec(`UPDATE habitaciones SET inquilino_id = ?, actualizado_en = CURRENT_TIMESTAMP WHERE id = ?`, inquilinoID, id)
	if err != nil {
		return domain.Habitacion{}, fmt.Errorf("asignando ocupante de la habitación %d: %w", id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return domain.Habitacion{}, fmt.Errorf("comprobando asignación de ocupante de la habitación %d: %w", id, err)
	}
	if n == 0 {
		return domain.Habitacion{}, ErrNotFound
	}
	return r.Get(id)
}

// Delete borra una habitación. Devuelve ErrNotFound si el id no existe.
func (r *HabitacionesRepo) Delete(id int64) error {
	res, err := r.db.Exec(`DELETE FROM habitaciones WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("borrando habitación %d: %w", id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("comprobando borrado de habitación %d: %w", id, err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
