package sqlite

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/naparito/Arantxator-flat-admin/internal/domain"
)

// NotificacionesRepo da acceso a la ÚNICA tabla que el centro de
// notificaciones persiste: notificaciones_leidas, el registro de qué avisos
// ha marcado el usuario como leídos. Los avisos en sí no se guardan: se
// evalúan al vuelo (ver internal/domain/notificacion.go y
// internal/httpapi/dashboard.go).
type NotificacionesRepo struct {
	db *sql.DB
}

func NewNotificacionesRepo(db *sql.DB) *NotificacionesRepo {
	return &NotificacionesRepo{db: db}
}

// MarcarLeida registra (o refresca) que el aviso identificado por `clave` está
// leído. Es idempotente: marcar dos veces el mismo aviso no falla, solo
// actualiza leida_en. `clave` es la identidad determinista
// domain.ClaveNotificacion(tipo, entidadTipo, entidadID).
func (r *NotificacionesRepo) MarcarLeida(clave string, tipo domain.TipoNotificacion, entidadTipo domain.EntidadTipo, entidadID int64) error {
	_, err := r.db.Exec(`
		INSERT INTO notificaciones_leidas (clave, tipo, entidad_tipo, entidad_id, leida_en)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(clave) DO UPDATE SET leida_en = CURRENT_TIMESTAMP`,
		clave, string(tipo), string(entidadTipo), entidadID,
	)
	if err != nil {
		return fmt.Errorf("marcando la notificación %q como leída: %w", clave, err)
	}
	return nil
}

// ClavesLeidas devuelve el conjunto de claves de avisos marcados como leídos,
// con su fecha de lectura. El motor de reglas lo usa para poner Leida = true
// en los avisos que siguen activos y descontarlos del contador.
func (r *NotificacionesRepo) ClavesLeidas() (map[string]time.Time, error) {
	rows, err := r.db.Query(`SELECT clave, leida_en FROM notificaciones_leidas`)
	if err != nil {
		return nil, fmt.Errorf("leyendo notificaciones leídas: %w", err)
	}
	defer rows.Close()

	leidas := map[string]time.Time{}
	for rows.Next() {
		var (
			clave string
			en    time.Time
		)
		if err := rows.Scan(&clave, &en); err != nil {
			return nil, fmt.Errorf("leyendo fila de notificación leída: %w", err)
		}
		leidas[clave] = en
	}
	return leidas, rows.Err()
}
