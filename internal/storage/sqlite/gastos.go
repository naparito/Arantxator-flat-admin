package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/naparito/Arantxator-flat-admin/internal/domain"
)

// GastosRepo da acceso a la tabla gastos (facturas por inmueble). El reparto
// entre inquilinos vive en repartos_gasto (ver RepartosRepo); el recibo
// individual de cada factura se calcula al vuelo (domain.CalcularRecibo), no
// se persiste.
type GastosRepo struct {
	db *sql.DB
}

func NewGastosRepo(db *sql.DB) *GastosRepo {
	return &GastosRepo{db: db}
}

const gastoColumns = `
	id, inmueble_id, tipo, periodicidad, importe, fecha_emision, fecha_vencimiento,
	proveedor, estado_pago, fecha_pago, metodo_pago, creado_en, actualizado_en
`

func scanGasto(row interface{ Scan(...any) error }) (domain.Gasto, error) {
	var (
		g            domain.Gasto
		periodicidad sql.NullString
	)
	err := row.Scan(
		&g.ID, &g.InmuebleID, &g.Tipo, &periodicidad, &g.Importe, &g.FechaEmision, &g.FechaVencimiento,
		&g.Proveedor, &g.EstadoPago, &g.FechaPago, &g.MetodoPago, &g.CreadoEn, &g.ActualizadoEn,
	)
	if err != nil {
		return domain.Gasto{}, err
	}
	g.Periodicidad = domain.Periodicidad(periodicidad.String)
	return g, nil
}

// Create inserta una factura nueva. estado_pago entra como "pendiente" o
// "pagado" (nunca "vencido": ese estado se deriva al leer); si llega
// "pagado" sin fecha_pago se sella con la fecha de hoy.
func (r *GastosRepo) Create(g domain.Gasto) (domain.Gasto, error) {
	res, err := r.db.Exec(`
		INSERT INTO gastos (
			inmueble_id, tipo, periodicidad, importe, fecha_emision, fecha_vencimiento,
			proveedor, estado_pago, fecha_pago, metodo_pago
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		g.InmuebleID, g.Tipo, nullIfEmpty(string(g.Periodicidad)), g.Importe, &g.FechaEmision, g.FechaVencimiento,
		g.Proveedor, estadoPagoOrDefault(g.EstadoPago), g.FechaPago, g.MetodoPago,
	)
	if err != nil {
		return domain.Gasto{}, fmt.Errorf("insertando gasto: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return domain.Gasto{}, fmt.Errorf("obteniendo id del gasto creado: %w", err)
	}
	return r.Get(id)
}

// Get devuelve una factura por id, o ErrNotFound.
func (r *GastosRepo) Get(id int64) (domain.Gasto, error) {
	row := r.db.QueryRow(`SELECT `+gastoColumns+` FROM gastos WHERE id = ?`, id)
	g, err := scanGasto(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Gasto{}, ErrNotFound
	}
	if err != nil {
		return domain.Gasto{}, fmt.Errorf("leyendo gasto %d: %w", id, err)
	}
	return g, nil
}

// ListByInmueble devuelve las facturas de un inmueble, las más recientes
// (por fecha de emisión) primero.
func (r *GastosRepo) ListByInmueble(inmuebleID int64) ([]domain.Gasto, error) {
	rows, err := r.db.Query(`
		SELECT `+gastoColumns+` FROM gastos
		WHERE inmueble_id = ?
		ORDER BY fecha_emision DESC, id DESC`, inmuebleID)
	if err != nil {
		return nil, fmt.Errorf("listando gastos del inmueble %d: %w", inmuebleID, err)
	}
	defer rows.Close()

	gastos := []domain.Gasto{}
	for rows.Next() {
		g, err := scanGasto(rows)
		if err != nil {
			return nil, fmt.Errorf("leyendo fila de gasto: %w", err)
		}
		gastos = append(gastos, g)
	}
	return gastos, rows.Err()
}

// List devuelve todas las facturas de toda la cartera, las más recientes
// (por fecha de emisión) primero. Alimenta el resumen del dashboard y el
// motor de reglas de notificación (§8), que necesitan cruzar gastos de todos
// los inmuebles sin ir inmueble por inmueble.
func (r *GastosRepo) List() ([]domain.Gasto, error) {
	rows, err := r.db.Query(`
		SELECT ` + gastoColumns + ` FROM gastos
		ORDER BY fecha_emision DESC, id DESC`)
	if err != nil {
		return nil, fmt.Errorf("listando gastos: %w", err)
	}
	defer rows.Close()

	gastos := []domain.Gasto{}
	for rows.Next() {
		g, err := scanGasto(rows)
		if err != nil {
			return nil, fmt.Errorf("leyendo fila de gasto: %w", err)
		}
		gastos = append(gastos, g)
	}
	return gastos, rows.Err()
}

// Update sobrescribe los campos editables de una factura (incluye marcarla
// como pagada). Si pasa a "pagado" sin fecha_pago se sella con hoy; si vuelve
// a "pendiente" se limpia la fecha_pago.
func (r *GastosRepo) Update(id int64, g domain.Gasto) (domain.Gasto, error) {
	estado := estadoPagoOrDefault(g.EstadoPago)
	fechaPago := g.FechaPago
	if estado == domain.PagoPagado && fechaPago == nil {
		hoy := domain.Fecha(time.Now())
		fechaPago = &hoy
	}
	if estado == domain.PagoPendiente {
		fechaPago = nil
	}

	res, err := r.db.Exec(`
		UPDATE gastos SET
			tipo = ?, periodicidad = ?, importe = ?, fecha_emision = ?, fecha_vencimiento = ?,
			proveedor = ?, estado_pago = ?, fecha_pago = ?, metodo_pago = ?,
			actualizado_en = CURRENT_TIMESTAMP
		WHERE id = ?`,
		g.Tipo, nullIfEmpty(string(g.Periodicidad)), g.Importe, &g.FechaEmision, g.FechaVencimiento,
		g.Proveedor, estado, fechaPago, g.MetodoPago,
		id,
	)
	if err != nil {
		return domain.Gasto{}, fmt.Errorf("actualizando gasto %d: %w", id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return domain.Gasto{}, fmt.Errorf("comprobando actualización de gasto %d: %w", id, err)
	}
	if n == 0 {
		return domain.Gasto{}, ErrNotFound
	}
	return r.Get(id)
}

// SumaImporteEnPeriodo suma el importe de las facturas de un inmueble cuya
// fecha de emisión cae en [desde, hasta) — la base del término "gastos" de
// la rentabilidad neta. El filtro por fecha se hace en Go (como el cálculo
// del % de ocupación de contratos) para no depender de en qué formato deja
// el driver las columnas DATE al compararlas en SQL.
func (r *GastosRepo) SumaImporteEnPeriodo(inmuebleID int64, desde, hasta time.Time) (float64, error) {
	gastos, err := r.ListByInmueble(inmuebleID)
	if err != nil {
		return 0, err
	}
	d, h := soloDia(desde), soloDia(hasta)
	total := 0.0
	for _, g := range gastos {
		em := soloDia(time.Time(g.FechaEmision))
		if !em.Before(d) && em.Before(h) {
			total += g.Importe
		}
	}
	return total, nil
}

// soloDia descarta la parte horaria para comparar por día natural.
func soloDia(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func estadoPagoOrDefault(e domain.EstadoPago) domain.EstadoPago {
	if e == domain.PagoPagado {
		return domain.PagoPagado
	}
	// "vencido" nunca se persiste: es un estado derivado al leer.
	return domain.PagoPendiente
}
