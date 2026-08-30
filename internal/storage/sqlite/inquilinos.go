package sqlite

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/naparito/Arantxator-flat-admin/internal/domain"
)

// InquilinosRepo da acceso a la tabla inquilinos.
type InquilinosRepo struct {
	db *sql.DB
}

func NewInquilinosRepo(db *sql.DB) *InquilinosRepo {
	return &InquilinosRepo{db: db}
}

const inquilinoColumns = `
	id, nombre_completo, documento_identidad, fecha_nacimiento, telefono, email, nacionalidad,
	contacto_emergencia_nombre, contacto_emergencia_telefono, iban, creado_en, actualizado_en
`

func scanInquilino(row interface{ Scan(...any) error }) (domain.Inquilino, error) {
	var i domain.Inquilino
	err := row.Scan(
		&i.ID, &i.NombreCompleto, &i.DocumentoIdentidad, &i.FechaNacimiento, &i.Telefono, &i.Email, &i.Nacionalidad,
		&i.ContactoEmergenciaNombre, &i.ContactoEmergenciaTelefono, &i.IBAN, &i.CreadoEn, &i.ActualizadoEn,
	)
	return i, err
}

// Create inserta un inquilino nuevo y devuelve el registro tal como queda
// persistido (con id y timestamps).
func (r *InquilinosRepo) Create(i domain.Inquilino) (domain.Inquilino, error) {
	res, err := r.db.Exec(`
		INSERT INTO inquilinos (
			nombre_completo, documento_identidad, fecha_nacimiento, telefono, email, nacionalidad,
			contacto_emergencia_nombre, contacto_emergencia_telefono, iban
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		i.NombreCompleto, i.DocumentoIdentidad, i.FechaNacimiento, i.Telefono, i.Email, i.Nacionalidad,
		i.ContactoEmergenciaNombre, i.ContactoEmergenciaTelefono, i.IBAN,
	)
	if err != nil {
		return domain.Inquilino{}, fmt.Errorf("insertando inquilino: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return domain.Inquilino{}, fmt.Errorf("obteniendo id del inquilino creado: %w", err)
	}
	return r.Get(id)
}

// Get devuelve un inquilino por id, o ErrNotFound si no existe.
func (r *InquilinosRepo) Get(id int64) (domain.Inquilino, error) {
	row := r.db.QueryRow(`SELECT `+inquilinoColumns+` FROM inquilinos WHERE id = ?`, id)
	i, err := scanInquilino(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Inquilino{}, ErrNotFound
	}
	if err != nil {
		return domain.Inquilino{}, fmt.Errorf("leyendo inquilino %d: %w", id, err)
	}
	return i, nil
}

// List devuelve todos los inquilinos, ordenados por nombre. La búsqueda por
// nombre/documento se resuelve en la GUI sobre este listado completo, igual
// que el filtro por estado de Inmuebles.
func (r *InquilinosRepo) List() ([]domain.Inquilino, error) {
	rows, err := r.db.Query(`SELECT ` + inquilinoColumns + ` FROM inquilinos ORDER BY nombre_completo`)
	if err != nil {
		return nil, fmt.Errorf("listando inquilinos: %w", err)
	}
	defer rows.Close()

	inquilinos := []domain.Inquilino{}
	for rows.Next() {
		i, err := scanInquilino(rows)
		if err != nil {
			return nil, fmt.Errorf("leyendo fila de inquilino: %w", err)
		}
		inquilinos = append(inquilinos, i)
	}
	return inquilinos, rows.Err()
}

// Update sobrescribe los campos editables de un inquilino existente y
// refresca actualizado_en. Devuelve ErrNotFound si el id no existe.
func (r *InquilinosRepo) Update(id int64, i domain.Inquilino) (domain.Inquilino, error) {
	res, err := r.db.Exec(`
		UPDATE inquilinos SET
			nombre_completo = ?, documento_identidad = ?, fecha_nacimiento = ?, telefono = ?, email = ?, nacionalidad = ?,
			contacto_emergencia_nombre = ?, contacto_emergencia_telefono = ?, iban = ?,
			actualizado_en = CURRENT_TIMESTAMP
		WHERE id = ?`,
		i.NombreCompleto, i.DocumentoIdentidad, i.FechaNacimiento, i.Telefono, i.Email, i.Nacionalidad,
		i.ContactoEmergenciaNombre, i.ContactoEmergenciaTelefono, i.IBAN,
		id,
	)
	if err != nil {
		return domain.Inquilino{}, fmt.Errorf("actualizando inquilino %d: %w", id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return domain.Inquilino{}, fmt.Errorf("comprobando actualización de inquilino %d: %w", id, err)
	}
	if n == 0 {
		return domain.Inquilino{}, ErrNotFound
	}
	return r.Get(id)
}
