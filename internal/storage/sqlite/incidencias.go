package sqlite

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/naparito/Arantxator-flat-admin/internal/domain"
)

// IncidenciasRepo da acceso a la tabla incidencias (submódulo de Inmuebles)
// y a su historial fechado en incidencia_eventos: el alta, cada cambio de
// estado del flujo (abierta → en proceso → esperando proveedor → resuelta →
// cerrada) y los comentarios de seguimiento.
type IncidenciasRepo struct {
	db *sql.DB
}

func NewIncidenciasRepo(db *sql.DB) *IncidenciasRepo {
	return &IncidenciasRepo{db: db}
}

const incidenciaColumns = `
	id, inmueble_id, titulo, descripcion, categoria, prioridad, origen, estado,
	proveedor_nombre, proveedor_contacto, coste, coste_a_cargo_de,
	fecha_apertura, fecha_cierre, creado_en, actualizado_en
`

func scanIncidencia(row interface{ Scan(...any) error }) (domain.Incidencia, error) {
	var (
		i      domain.Incidencia
		origen sql.NullString
		cargo  sql.NullString
	)
	err := row.Scan(
		&i.ID, &i.InmuebleID, &i.Titulo, &i.Descripcion, &i.Categoria, &i.Prioridad, &origen, &i.Estado,
		&i.ProveedorNombre, &i.ProveedorContacto, &i.Coste, &cargo,
		&i.FechaApertura, &i.FechaCierre, &i.CreadoEn, &i.ActualizadoEn,
	)
	if err != nil {
		return domain.Incidencia{}, err
	}
	i.Origen = domain.OrigenIncidencia(origen.String)
	i.CosteACargoDe = domain.CosteACargoDe(cargo.String)
	i.Eventos = []domain.IncidenciaEvento{}
	return i, nil
}

// nullIfEmpty devuelve nil (para persistir NULL) cuando la cadena está
// vacía. coste_a_cargo_de tiene un CHECK que no admite ”, así que ahí sí
// hay que guardar NULL de verdad, no una cadena vacía.
func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// Create inserta una incidencia nueva. El estado siempre arranca en
// "abierta" (el flujo se recorre luego a mano con Update); fecha_apertura la
// pone SQLite con CURRENT_TIMESTAMP. Se registra el evento de alta.
func (r *IncidenciasRepo) Create(inc domain.Incidencia) (domain.Incidencia, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return domain.Incidencia{}, fmt.Errorf("iniciando transacción para crear incidencia: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.Exec(`
		INSERT INTO incidencias (
			inmueble_id, titulo, descripcion, categoria, prioridad, origen, estado,
			proveedor_nombre, proveedor_contacto, coste, coste_a_cargo_de
		) VALUES (?, ?, ?, ?, ?, ?, 'abierta', ?, ?, ?, ?)`,
		inc.InmuebleID, inc.Titulo, inc.Descripcion, inc.Categoria,
		prioridadOrDefault(inc.Prioridad), nullIfEmpty(string(inc.Origen)),
		inc.ProveedorNombre, inc.ProveedorContacto, inc.Coste, nullIfEmpty(string(inc.CosteACargoDe)),
	)
	if err != nil {
		return domain.Incidencia{}, fmt.Errorf("insertando incidencia: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return domain.Incidencia{}, fmt.Errorf("obteniendo id de la incidencia creada: %w", err)
	}

	if _, err := tx.Exec(`
		INSERT INTO incidencia_eventos (incidencia_id, tipo, estado_nuevo)
		VALUES (?, 'alta', 'abierta')`, id); err != nil {
		return domain.Incidencia{}, fmt.Errorf("registrando el alta de la incidencia %d: %w", id, err)
	}

	if inc.Comentario != "" {
		if _, err := tx.Exec(`
			INSERT INTO incidencia_eventos (incidencia_id, tipo, comentario)
			VALUES (?, 'comentario', ?)`, id, inc.Comentario); err != nil {
			return domain.Incidencia{}, fmt.Errorf("registrando el comentario de alta de la incidencia %d: %w", id, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return domain.Incidencia{}, fmt.Errorf("confirmando creación de la incidencia: %w", err)
	}
	return r.Get(id)
}

// Get devuelve una incidencia por id, con su historial de eventos, o
// ErrNotFound.
func (r *IncidenciasRepo) Get(id int64) (domain.Incidencia, error) {
	row := r.db.QueryRow(`SELECT `+incidenciaColumns+` FROM incidencias WHERE id = ?`, id)
	inc, err := scanIncidencia(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Incidencia{}, ErrNotFound
	}
	if err != nil {
		return domain.Incidencia{}, fmt.Errorf("leyendo incidencia %d: %w", id, err)
	}
	eventos, err := r.eventos(id)
	if err != nil {
		return domain.Incidencia{}, err
	}
	inc.Eventos = eventos
	return inc, nil
}

// ListByInmueble devuelve las incidencias de un inmueble, las más recientes
// (por fecha de apertura) primero, cada una con su historial de eventos.
func (r *IncidenciasRepo) ListByInmueble(inmuebleID int64) ([]domain.Incidencia, error) {
	rows, err := r.db.Query(`
		SELECT `+incidenciaColumns+` FROM incidencias
		WHERE inmueble_id = ?
		ORDER BY fecha_apertura DESC, id DESC`, inmuebleID)
	if err != nil {
		return nil, fmt.Errorf("listando incidencias del inmueble %d: %w", inmuebleID, err)
	}
	defer rows.Close()

	incidencias := []domain.Incidencia{}
	for rows.Next() {
		inc, err := scanIncidencia(rows)
		if err != nil {
			return nil, fmt.Errorf("leyendo fila de incidencia: %w", err)
		}
		incidencias = append(incidencias, inc)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return r.rellenarEventos(inmuebleID, incidencias)
}

// Update sobrescribe los campos editables de una incidencia. Si el estado
// cambia respecto al guardado, deja constancia del cambio con su fecha en
// incidencia_eventos y ajusta fecha_cierre (se fija al pasar a "cerrada", se
// limpia al reabrir). La validez de la transición se comprueba antes, en la
// capa HTTP. Un Comentario no vacío se añade como evento de seguimiento.
func (r *IncidenciasRepo) Update(id int64, inc domain.Incidencia) (domain.Incidencia, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return domain.Incidencia{}, fmt.Errorf("iniciando transacción para actualizar incidencia: %w", err)
	}
	defer tx.Rollback()

	var estadoAnterior domain.EstadoIncidencia
	err = tx.QueryRow(`SELECT estado FROM incidencias WHERE id = ?`, id).Scan(&estadoAnterior)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Incidencia{}, ErrNotFound
	}
	if err != nil {
		return domain.Incidencia{}, fmt.Errorf("leyendo estado actual de la incidencia %d: %w", id, err)
	}

	nuevoEstado := inc.Estado
	if nuevoEstado == "" {
		nuevoEstado = estadoAnterior
	}

	// fecha_cierre: se sella al cerrar, se limpia si la incidencia se reabre.
	fechaCierreExpr := "fecha_cierre"
	if nuevoEstado.EsFinal() {
		fechaCierreExpr = "COALESCE(fecha_cierre, CURRENT_TIMESTAMP)"
	} else {
		fechaCierreExpr = "NULL"
	}

	_, err = tx.Exec(`
		UPDATE incidencias SET
			titulo = ?, descripcion = ?, categoria = ?, prioridad = ?, origen = ?, estado = ?,
			proveedor_nombre = ?, proveedor_contacto = ?, coste = ?, coste_a_cargo_de = ?,
			fecha_cierre = `+fechaCierreExpr+`,
			actualizado_en = CURRENT_TIMESTAMP
		WHERE id = ?`,
		inc.Titulo, inc.Descripcion, inc.Categoria, prioridadOrDefault(inc.Prioridad),
		nullIfEmpty(string(inc.Origen)), nuevoEstado,
		inc.ProveedorNombre, inc.ProveedorContacto, inc.Coste, nullIfEmpty(string(inc.CosteACargoDe)),
		id,
	)
	if err != nil {
		return domain.Incidencia{}, fmt.Errorf("actualizando incidencia %d: %w", id, err)
	}

	if nuevoEstado != estadoAnterior {
		if _, err := tx.Exec(`
			INSERT INTO incidencia_eventos (incidencia_id, tipo, estado_anterior, estado_nuevo)
			VALUES (?, 'cambio_estado', ?, ?)`, id, estadoAnterior, nuevoEstado); err != nil {
			return domain.Incidencia{}, fmt.Errorf("registrando el cambio de estado de la incidencia %d: %w", id, err)
		}
	}

	if inc.Comentario != "" {
		if _, err := tx.Exec(`
			INSERT INTO incidencia_eventos (incidencia_id, tipo, comentario)
			VALUES (?, 'comentario', ?)`, id, inc.Comentario); err != nil {
			return domain.Incidencia{}, fmt.Errorf("registrando el comentario de la incidencia %d: %w", id, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return domain.Incidencia{}, fmt.Errorf("confirmando actualización de la incidencia: %w", err)
	}
	return r.Get(id)
}

// CuentaAbiertasByInmueble devuelve cuántas incidencias del inmueble no
// están en un estado final ("cerrada"): es el número que muestra el badge
// del tab Incidencias.
func (r *IncidenciasRepo) CuentaAbiertasByInmueble(inmuebleID int64) (int, error) {
	var n int
	err := r.db.QueryRow(`
		SELECT COUNT(*) FROM incidencias
		WHERE inmueble_id = ? AND estado != 'cerrada'`, inmuebleID).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("contando incidencias abiertas del inmueble %d: %w", inmuebleID, err)
	}
	return n, nil
}

func (r *IncidenciasRepo) eventos(incidenciaID int64) ([]domain.IncidenciaEvento, error) {
	rows, err := r.db.Query(`
		SELECT id, incidencia_id, tipo, estado_anterior, estado_nuevo, comentario, creado_en
		FROM incidencia_eventos WHERE incidencia_id = ? ORDER BY id`, incidenciaID)
	if err != nil {
		return nil, fmt.Errorf("leyendo eventos de la incidencia %d: %w", incidenciaID, err)
	}
	defer rows.Close()
	return scanEventos(rows)
}

// rellenarEventos completa el historial de una lista de incidencias de un
// mismo inmueble con una sola consulta, igual que rellenarCoArrendatarios
// en contratos.go.
func (r *IncidenciasRepo) rellenarEventos(inmuebleID int64, incidencias []domain.Incidencia) ([]domain.Incidencia, error) {
	if len(incidencias) == 0 {
		return incidencias, nil
	}
	rows, err := r.db.Query(`
		SELECT e.id, e.incidencia_id, e.tipo, e.estado_anterior, e.estado_nuevo, e.comentario, e.creado_en
		FROM incidencia_eventos e
		JOIN incidencias i ON i.id = e.incidencia_id
		WHERE i.inmueble_id = ?
		ORDER BY e.id`, inmuebleID)
	if err != nil {
		return nil, fmt.Errorf("leyendo eventos de las incidencias: %w", err)
	}
	defer rows.Close()

	todos, err := scanEventos(rows)
	if err != nil {
		return nil, err
	}
	porIncidencia := map[int64][]domain.IncidenciaEvento{}
	for _, e := range todos {
		porIncidencia[e.IncidenciaID] = append(porIncidencia[e.IncidenciaID], e)
	}
	for idx := range incidencias {
		ev := porIncidencia[incidencias[idx].ID]
		if ev == nil {
			ev = []domain.IncidenciaEvento{}
		}
		incidencias[idx].Eventos = ev
	}
	return incidencias, nil
}

func scanEventos(rows *sql.Rows) ([]domain.IncidenciaEvento, error) {
	eventos := []domain.IncidenciaEvento{}
	for rows.Next() {
		var (
			e        domain.IncidenciaEvento
			anterior sql.NullString
			nuevo    sql.NullString
			coment   sql.NullString
		)
		if err := rows.Scan(&e.ID, &e.IncidenciaID, &e.Tipo, &anterior, &nuevo, &coment, &e.CreadoEn); err != nil {
			return nil, fmt.Errorf("leyendo fila de evento de incidencia: %w", err)
		}
		e.EstadoAnterior = domain.EstadoIncidencia(anterior.String)
		e.EstadoNuevo = domain.EstadoIncidencia(nuevo.String)
		e.Comentario = coment.String
		eventos = append(eventos, e)
	}
	return eventos, rows.Err()
}

func prioridadOrDefault(p domain.PrioridadIncidencia) domain.PrioridadIncidencia {
	if p == "" {
		return domain.PrioridadMedia
	}
	return p
}
