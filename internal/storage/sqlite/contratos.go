package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/naparito/Arantxator-flat-admin/internal/domain"
)

// ContratosRepo da acceso a la tabla contratos y a su relación N:N con
// inquilinos (contrato_inquilino). En un inmueble no compartido el contrato
// es del inmueble entero (habitacion_id NULL); en uno compartido, de una
// habitación concreta (habitacion_id no nulo).
type ContratosRepo struct {
	db *sql.DB
}

func NewContratosRepo(db *sql.DB) *ContratosRepo {
	return &ContratosRepo{db: db}
}

const contratoColumns = `
	id, inmueble_id, habitacion_id, fecha_firma, fecha_inicio, fecha_fin,
	arrendador_persona_juridica, renta_mensual, dia_pago, indice_actualizacion,
	proxima_revision_renta, fianza_importe, fianza_estado, fianza_fecha_deposito,
	estado, motivo_baja, creado_en, actualizado_en
`

func scanContrato(row interface{ Scan(...any) error }) (domain.Contrato, error) {
	var c domain.Contrato
	err := row.Scan(
		&c.ID, &c.InmuebleID, &c.HabitacionID, &c.FechaFirma, &c.FechaInicio, &c.FechaFin,
		&c.ArrendadorPersonaJuridica, &c.RentaMensual, &c.DiaPago, &c.IndiceActualizacion,
		&c.ProximaRevisionRenta, &c.FianzaImporte, &c.FianzaEstado, &c.FianzaFechaDeposito,
		&c.Estado, &c.MotivoBaja, &c.CreadoEn, &c.ActualizadoEn,
	)
	return c, err
}

// Create inserta un contrato nuevo con sus co-arrendatarios y, si el inmueble
// no es compartido y el contrato queda vigente, pone el inmueble en
// "alquilado" automáticamente. Todo en una única transacción.
func (r *ContratosRepo) Create(c domain.Contrato) (domain.Contrato, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return domain.Contrato{}, fmt.Errorf("iniciando transacción para crear contrato: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.Exec(`
		INSERT INTO contratos (
			inmueble_id, habitacion_id, fecha_firma, fecha_inicio, fecha_fin,
			arrendador_persona_juridica, renta_mensual, dia_pago, indice_actualizacion,
			proxima_revision_renta, fianza_importe, fianza_estado, fianza_fecha_deposito,
			estado, motivo_baja
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		c.InmuebleID, c.HabitacionID, &c.FechaFirma, &c.FechaInicio, &c.FechaFin,
		c.ArrendadorPersonaJuridica, c.RentaMensual, c.DiaPago, indiceOrDefault(c.IndiceActualizacion),
		c.ProximaRevisionRenta, c.FianzaImporte, fianzaEstadoOrDefault(c.FianzaEstado), c.FianzaFechaDeposito,
		estadoContratoOrDefault(c.Estado), c.MotivoBaja,
	)
	if err != nil {
		return domain.Contrato{}, fmt.Errorf("insertando contrato: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return domain.Contrato{}, fmt.Errorf("obteniendo id del contrato creado: %w", err)
	}

	if err := reemplazarCoArrendatarios(tx, id, c.InquilinoIDs); err != nil {
		return domain.Contrato{}, err
	}

	c.ID = id
	if err := sincronizarEstadoInmueble(tx, c); err != nil {
		return domain.Contrato{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.Contrato{}, fmt.Errorf("confirmando creación del contrato: %w", err)
	}
	return r.Get(id)
}

// Get devuelve un contrato por id (con sus inquilinoIds), o ErrNotFound.
func (r *ContratosRepo) Get(id int64) (domain.Contrato, error) {
	row := r.db.QueryRow(`SELECT `+contratoColumns+` FROM contratos WHERE id = ?`, id)
	c, err := scanContrato(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Contrato{}, ErrNotFound
	}
	if err != nil {
		return domain.Contrato{}, fmt.Errorf("leyendo contrato %d: %w", id, err)
	}
	ids, err := r.inquilinoIDs(id)
	if err != nil {
		return domain.Contrato{}, err
	}
	c.InquilinoIDs = ids
	return c, nil
}

// List devuelve todos los contratos, más recientes primero (por fecha de
// inicio), cada uno con sus co-arrendatarios.
func (r *ContratosRepo) List() ([]domain.Contrato, error) {
	rows, err := r.db.Query(`SELECT ` + contratoColumns + ` FROM contratos ORDER BY fecha_inicio DESC, id DESC`)
	if err != nil {
		return nil, fmt.Errorf("listando contratos: %w", err)
	}
	defer rows.Close()

	contratos := []domain.Contrato{}
	for rows.Next() {
		c, err := scanContrato(rows)
		if err != nil {
			return nil, fmt.Errorf("leyendo fila de contrato: %w", err)
		}
		contratos = append(contratos, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return r.rellenarCoArrendatarios(contratos)
}

// ListByInquilino devuelve los contratos en los que el inquilino figura como
// co-arrendatario, más recientes primero. Alimenta el histórico de su ficha.
func (r *ContratosRepo) ListByInquilino(inquilinoID int64) ([]domain.Contrato, error) {
	rows, err := r.db.Query(`
		SELECT `+contratoColumns+` FROM contratos
		WHERE id IN (SELECT contrato_id FROM contrato_inquilino WHERE inquilino_id = ?)
		ORDER BY fecha_inicio DESC, id DESC`, inquilinoID)
	if err != nil {
		return nil, fmt.Errorf("listando contratos del inquilino %d: %w", inquilinoID, err)
	}
	defer rows.Close()

	contratos := []domain.Contrato{}
	for rows.Next() {
		c, err := scanContrato(rows)
		if err != nil {
			return nil, fmt.Errorf("leyendo fila de contrato: %w", err)
		}
		contratos = append(contratos, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return r.rellenarCoArrendatarios(contratos)
}

// Update sobrescribe los campos editables de un contrato, resincroniza sus
// co-arrendatarios y, si el inmueble no es compartido, ajusta su estado
// ("disponible" al rescindir, "alquilado" si vuelve a estar vigente).
func (r *ContratosRepo) Update(id int64, c domain.Contrato) (domain.Contrato, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return domain.Contrato{}, fmt.Errorf("iniciando transacción para actualizar contrato: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.Exec(`
		UPDATE contratos SET
			inmueble_id = ?, habitacion_id = ?, fecha_firma = ?, fecha_inicio = ?, fecha_fin = ?,
			arrendador_persona_juridica = ?, renta_mensual = ?, dia_pago = ?, indice_actualizacion = ?,
			proxima_revision_renta = ?, fianza_importe = ?, fianza_estado = ?, fianza_fecha_deposito = ?,
			estado = ?, motivo_baja = ?, actualizado_en = CURRENT_TIMESTAMP
		WHERE id = ?`,
		c.InmuebleID, c.HabitacionID, &c.FechaFirma, &c.FechaInicio, &c.FechaFin,
		c.ArrendadorPersonaJuridica, c.RentaMensual, c.DiaPago, indiceOrDefault(c.IndiceActualizacion),
		c.ProximaRevisionRenta, c.FianzaImporte, fianzaEstadoOrDefault(c.FianzaEstado), c.FianzaFechaDeposito,
		estadoContratoOrDefault(c.Estado), c.MotivoBaja,
		id,
	)
	if err != nil {
		return domain.Contrato{}, fmt.Errorf("actualizando contrato %d: %w", id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return domain.Contrato{}, fmt.Errorf("comprobando actualización de contrato %d: %w", id, err)
	}
	if n == 0 {
		return domain.Contrato{}, ErrNotFound
	}

	if err := reemplazarCoArrendatarios(tx, id, c.InquilinoIDs); err != nil {
		return domain.Contrato{}, err
	}

	c.ID = id
	if err := sincronizarEstadoInmueble(tx, c); err != nil {
		return domain.Contrato{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.Contrato{}, fmt.Errorf("confirmando actualización del contrato: %w", err)
	}
	return r.Get(id)
}

// TieneContratoVigenteEnAmbito indica si ya existe otro contrato vigente
// sobre el mismo ámbito: la habitación (si habitacionID no es nulo) o el
// inmueble completo (si lo es). excluirID permite ignorar el propio contrato
// al editarlo.
func (r *ContratosRepo) TieneContratoVigenteEnAmbito(inmuebleID int64, habitacionID *int64, excluirID int64, ref time.Time) (bool, error) {
	var (
		rows *sql.Rows
		err  error
	)
	if habitacionID != nil {
		rows, err = r.db.Query(`
			SELECT `+contratoColumns+` FROM contratos
			WHERE habitacion_id = ? AND id != ?`, *habitacionID, excluirID)
	} else {
		rows, err = r.db.Query(`
			SELECT `+contratoColumns+` FROM contratos
			WHERE inmueble_id = ? AND habitacion_id IS NULL AND id != ?`, inmuebleID, excluirID)
	}
	if err != nil {
		return false, fmt.Errorf("comprobando solapamiento de contratos: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		c, err := scanContrato(rows)
		if err != nil {
			return false, fmt.Errorf("leyendo fila de contrato: %w", err)
		}
		if c.EstaVigente(ref) {
			return true, nil
		}
	}
	return false, rows.Err()
}

// OcupacionInmueble cuenta, para un inmueble compartido, cuántas de sus
// habitaciones tienen al menos un contrato vigente y el total de
// habitaciones. Es la base del % de ocupación que se muestra en el listado y
// la ficha de Inmuebles.
func (r *ContratosRepo) OcupacionInmueble(inmuebleID int64, ref time.Time) (ocupadas, totales int, err error) {
	if err = r.db.QueryRow(`SELECT COUNT(*) FROM habitaciones WHERE inmueble_id = ?`, inmuebleID).Scan(&totales); err != nil {
		return 0, 0, fmt.Errorf("contando habitaciones del inmueble %d: %w", inmuebleID, err)
	}

	rows, err := r.db.Query(`
		SELECT `+contratoColumns+` FROM contratos
		WHERE inmueble_id = ? AND habitacion_id IS NOT NULL`, inmuebleID)
	if err != nil {
		return 0, 0, fmt.Errorf("leyendo contratos del inmueble %d: %w", inmuebleID, err)
	}
	defer rows.Close()

	habitacionesOcupadas := map[int64]bool{}
	for rows.Next() {
		c, err := scanContrato(rows)
		if err != nil {
			return 0, 0, fmt.Errorf("leyendo fila de contrato: %w", err)
		}
		if c.HabitacionID != nil && c.EstaVigente(ref) {
			habitacionesOcupadas[*c.HabitacionID] = true
		}
	}
	if err := rows.Err(); err != nil {
		return 0, 0, err
	}
	return len(habitacionesOcupadas), totales, nil
}

func (r *ContratosRepo) inquilinoIDs(contratoID int64) ([]int64, error) {
	rows, err := r.db.Query(`SELECT inquilino_id FROM contrato_inquilino WHERE contrato_id = ? ORDER BY inquilino_id`, contratoID)
	if err != nil {
		return nil, fmt.Errorf("leyendo co-arrendatarios del contrato %d: %w", contratoID, err)
	}
	defer rows.Close()

	ids := []int64{}
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("leyendo fila de co-arrendatario: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// rellenarCoArrendatarios completa el campo InquilinoIDs de una lista de
// contratos con una sola consulta.
func (r *ContratosRepo) rellenarCoArrendatarios(contratos []domain.Contrato) ([]domain.Contrato, error) {
	if len(contratos) == 0 {
		return contratos, nil
	}
	rows, err := r.db.Query(`SELECT contrato_id, inquilino_id FROM contrato_inquilino ORDER BY contrato_id, inquilino_id`)
	if err != nil {
		return nil, fmt.Errorf("leyendo co-arrendatarios: %w", err)
	}
	defer rows.Close()

	porContrato := map[int64][]int64{}
	for rows.Next() {
		var contratoID, inquilinoID int64
		if err := rows.Scan(&contratoID, &inquilinoID); err != nil {
			return nil, fmt.Errorf("leyendo fila de co-arrendatario: %w", err)
		}
		porContrato[contratoID] = append(porContrato[contratoID], inquilinoID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := range contratos {
		ids := porContrato[contratos[i].ID]
		if ids == nil {
			ids = []int64{}
		}
		contratos[i].InquilinoIDs = ids
	}
	return contratos, nil
}

func reemplazarCoArrendatarios(tx *sql.Tx, contratoID int64, inquilinoIDs []int64) error {
	if _, err := tx.Exec(`DELETE FROM contrato_inquilino WHERE contrato_id = ?`, contratoID); err != nil {
		return fmt.Errorf("limpiando co-arrendatarios del contrato %d: %w", contratoID, err)
	}
	for _, inquilinoID := range inquilinoIDs {
		if _, err := tx.Exec(`INSERT INTO contrato_inquilino (contrato_id, inquilino_id) VALUES (?, ?)`, contratoID, inquilinoID); err != nil {
			return fmt.Errorf("vinculando inquilino %d al contrato %d: %w", inquilinoID, contratoID, err)
		}
	}
	return nil
}

// sincronizarEstadoInmueble ajusta Inmueble.estado SOLO para inmuebles no
// compartidos (el guard `compartido = 0` lo garantiza): "alquilado" si el
// contrato queda vigente, "disponible" si se rescinde. En inmuebles
// compartidos el estado no se toca (lo lleva el % de ocupación).
func sincronizarEstadoInmueble(tx *sql.Tx, c domain.Contrato) error {
	if c.HabitacionID != nil {
		return nil // inmueble compartido: no se toca el estado
	}
	nuevo := domain.InmuebleDisponible
	if c.EstaVigente(time.Now()) {
		nuevo = domain.InmuebleAlquilado
	}
	if _, err := tx.Exec(`
		UPDATE inmuebles SET estado = ?, actualizado_en = CURRENT_TIMESTAMP
		WHERE id = ? AND compartido = 0`, nuevo, c.InmuebleID); err != nil {
		return fmt.Errorf("sincronizando estado del inmueble %d: %w", c.InmuebleID, err)
	}
	return nil
}

func indiceOrDefault(indice string) string {
	if indice == "" {
		return "IRAV"
	}
	return indice
}

func fianzaEstadoOrDefault(e domain.EstadoFianza) domain.EstadoFianza {
	if e == "" {
		return domain.FianzaPendiente
	}
	return e
}

func estadoContratoOrDefault(e domain.EstadoContrato) domain.EstadoContrato {
	if e == "" {
		return domain.ContratoActivo
	}
	return e
}
