package sqlite

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/naparito/Arantxator-flat-admin/internal/domain"
)

// ErrNotFound se devuelve cuando no existe un registro con el id pedido.
var ErrNotFound = errors.New("registro no encontrado")

// InmueblesRepo da acceso a la tabla inmuebles.
type InmueblesRepo struct {
	db *sql.DB
}

func NewInmueblesRepo(db *sql.DB) *InmueblesRepo {
	return &InmueblesRepo{db: db}
}

const inmuebleColumns = `
	id, nombre, direccion, referencia_catastral, codigo_postal, ciudad, provincia,
	tipo, m2_construidos, m2_utiles, num_habitaciones, num_banos, planta,
	ascensor, amueblado, anio_construccion,
	certificado_energetico_letra, certificado_energetico_caducidad, estado, compartido,
	luz_compania, luz_numero_contrato, luz_titular,
	agua_compania, agua_numero_contrato, agua_titular,
	gas_compania, gas_numero_contrato, gas_titular,
	internet_compania, internet_numero_contrato, internet_titular,
	creado_en, actualizado_en
`

func scanInmueble(row interface{ Scan(...any) error }) (domain.Inmueble, error) {
	var m domain.Inmueble
	err := row.Scan(
		&m.ID, &m.Nombre, &m.Direccion, &m.ReferenciaCatastral, &m.CodigoPostal, &m.Ciudad, &m.Provincia,
		&m.Tipo, &m.M2Construidos, &m.M2Utiles, &m.NumHabitaciones, &m.NumBanos, &m.Planta,
		&m.Ascensor, &m.Amueblado, &m.AnioConstruccion,
		&m.CertificadoEnergeticoLetra, &m.CertificadoEnergeticoCaducidad, &m.Estado, &m.Compartido,
		&m.Suministros.Luz.Compania, &m.Suministros.Luz.NumeroContrato, &m.Suministros.Luz.Titular,
		&m.Suministros.Agua.Compania, &m.Suministros.Agua.NumeroContrato, &m.Suministros.Agua.Titular,
		&m.Suministros.Gas.Compania, &m.Suministros.Gas.NumeroContrato, &m.Suministros.Gas.Titular,
		&m.Suministros.Internet.Compania, &m.Suministros.Internet.NumeroContrato, &m.Suministros.Internet.Titular,
		&m.CreadoEn, &m.ActualizadoEn,
	)
	return m, err
}

// Create inserta un inmueble nuevo y devuelve el registro tal como queda
// persistido (con id y timestamps).
func (r *InmueblesRepo) Create(m domain.Inmueble) (domain.Inmueble, error) {
	res, err := r.db.Exec(`
		INSERT INTO inmuebles (
			nombre, direccion, referencia_catastral, codigo_postal, ciudad, provincia,
			tipo, m2_construidos, m2_utiles, num_habitaciones, num_banos, planta,
			ascensor, amueblado, anio_construccion,
			certificado_energetico_letra, certificado_energetico_caducidad, estado, compartido,
			luz_compania, luz_numero_contrato, luz_titular,
			agua_compania, agua_numero_contrato, agua_titular,
			gas_compania, gas_numero_contrato, gas_titular,
			internet_compania, internet_numero_contrato, internet_titular
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		m.Nombre, m.Direccion, m.ReferenciaCatastral, m.CodigoPostal, m.Ciudad, m.Provincia,
		m.Tipo, m.M2Construidos, m.M2Utiles, m.NumHabitaciones, m.NumBanos, m.Planta,
		m.Ascensor, m.Amueblado, m.AnioConstruccion,
		m.CertificadoEnergeticoLetra, m.CertificadoEnergeticoCaducidad, estadoOrDefault(m.Estado), m.Compartido,
		m.Suministros.Luz.Compania, m.Suministros.Luz.NumeroContrato, m.Suministros.Luz.Titular,
		m.Suministros.Agua.Compania, m.Suministros.Agua.NumeroContrato, m.Suministros.Agua.Titular,
		m.Suministros.Gas.Compania, m.Suministros.Gas.NumeroContrato, m.Suministros.Gas.Titular,
		m.Suministros.Internet.Compania, m.Suministros.Internet.NumeroContrato, m.Suministros.Internet.Titular,
	)
	if err != nil {
		return domain.Inmueble{}, fmt.Errorf("insertando inmueble: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return domain.Inmueble{}, fmt.Errorf("obteniendo id del inmueble creado: %w", err)
	}
	return r.Get(id)
}

func estadoOrDefault(e domain.EstadoInmueble) domain.EstadoInmueble {
	if e == "" {
		return domain.InmuebleDisponible
	}
	return e
}

// Get devuelve un inmueble por id, o ErrNotFound si no existe.
func (r *InmueblesRepo) Get(id int64) (domain.Inmueble, error) {
	row := r.db.QueryRow(`SELECT `+inmuebleColumns+` FROM inmuebles WHERE id = ?`, id)
	m, err := scanInmueble(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Inmueble{}, ErrNotFound
	}
	if err != nil {
		return domain.Inmueble{}, fmt.Errorf("leyendo inmueble %d: %w", id, err)
	}
	return m, nil
}

// List devuelve los inmuebles, opcionalmente filtrados por estado. Un
// estado vacío no filtra.
func (r *InmueblesRepo) List(estado domain.EstadoInmueble) ([]domain.Inmueble, error) {
	query := `SELECT ` + inmuebleColumns + ` FROM inmuebles`
	args := []any{}
	if estado != "" {
		query += ` WHERE estado = ?`
		args = append(args, estado)
	}
	query += ` ORDER BY nombre`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("listando inmuebles: %w", err)
	}
	defer rows.Close()

	inmuebles := []domain.Inmueble{}
	for rows.Next() {
		m, err := scanInmueble(rows)
		if err != nil {
			return nil, fmt.Errorf("leyendo fila de inmueble: %w", err)
		}
		inmuebles = append(inmuebles, m)
	}
	return inmuebles, rows.Err()
}

// Update sobrescribe los campos editables de un inmueble existente y
// refresca actualizado_en. Devuelve ErrNotFound si el id no existe.
func (r *InmueblesRepo) Update(id int64, m domain.Inmueble) (domain.Inmueble, error) {
	res, err := r.db.Exec(`
		UPDATE inmuebles SET
			nombre = ?, direccion = ?, referencia_catastral = ?, codigo_postal = ?, ciudad = ?, provincia = ?,
			tipo = ?, m2_construidos = ?, m2_utiles = ?, num_habitaciones = ?, num_banos = ?, planta = ?,
			ascensor = ?, amueblado = ?, anio_construccion = ?,
			certificado_energetico_letra = ?, certificado_energetico_caducidad = ?, estado = ?, compartido = ?,
			luz_compania = ?, luz_numero_contrato = ?, luz_titular = ?,
			agua_compania = ?, agua_numero_contrato = ?, agua_titular = ?,
			gas_compania = ?, gas_numero_contrato = ?, gas_titular = ?,
			internet_compania = ?, internet_numero_contrato = ?, internet_titular = ?,
			actualizado_en = CURRENT_TIMESTAMP
		WHERE id = ?`,
		m.Nombre, m.Direccion, m.ReferenciaCatastral, m.CodigoPostal, m.Ciudad, m.Provincia,
		m.Tipo, m.M2Construidos, m.M2Utiles, m.NumHabitaciones, m.NumBanos, m.Planta,
		m.Ascensor, m.Amueblado, m.AnioConstruccion,
		m.CertificadoEnergeticoLetra, m.CertificadoEnergeticoCaducidad, estadoOrDefault(m.Estado), m.Compartido,
		m.Suministros.Luz.Compania, m.Suministros.Luz.NumeroContrato, m.Suministros.Luz.Titular,
		m.Suministros.Agua.Compania, m.Suministros.Agua.NumeroContrato, m.Suministros.Agua.Titular,
		m.Suministros.Gas.Compania, m.Suministros.Gas.NumeroContrato, m.Suministros.Gas.Titular,
		m.Suministros.Internet.Compania, m.Suministros.Internet.NumeroContrato, m.Suministros.Internet.Titular,
		id,
	)
	if err != nil {
		return domain.Inmueble{}, fmt.Errorf("actualizando inmueble %d: %w", id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return domain.Inmueble{}, fmt.Errorf("comprobando actualización de inmueble %d: %w", id, err)
	}
	if n == 0 {
		return domain.Inmueble{}, ErrNotFound
	}
	return r.Get(id)
}

// Archivar pone el inmueble como fuera de servicio sin borrar su historial
// (documentos, incidencias, etc. asociados).
func (r *InmueblesRepo) Archivar(id int64) error {
	res, err := r.db.Exec(`UPDATE inmuebles SET estado = ?, actualizado_en = CURRENT_TIMESTAMP WHERE id = ?`,
		domain.InmuebleFueraDeServicio, id)
	if err != nil {
		return fmt.Errorf("archivando inmueble %d: %w", id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("comprobando archivado de inmueble %d: %w", id, err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
