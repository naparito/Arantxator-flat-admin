package sqlite_test

import (
	"testing"

	"github.com/naparito/Arantxator-flat-admin/internal/domain"
	"github.com/naparito/Arantxator-flat-admin/internal/storage/sqlite"
)

func TestNotificacionesRepo_MarcarLeidaYConsultar(t *testing.T) {
	db := newTestDB(t)
	repo := sqlite.NewNotificacionesRepo(db)

	leidas, err := repo.ClavesLeidas()
	if err != nil {
		t.Fatalf("ClavesLeidas inicial: %v", err)
	}
	if len(leidas) != 0 {
		t.Fatalf("una base nueva no debería tener notificaciones leídas, tiene %d", len(leidas))
	}

	clave := domain.ClaveNotificacion(domain.NotifContratoPorVencer, domain.EntidadContrato, 42)
	if err := repo.MarcarLeida(clave, domain.NotifContratoPorVencer, domain.EntidadContrato, 42); err != nil {
		t.Fatalf("MarcarLeida: %v", err)
	}

	leidas, err = repo.ClavesLeidas()
	if err != nil {
		t.Fatalf("ClavesLeidas tras marcar: %v", err)
	}
	if _, ok := leidas[clave]; !ok || len(leidas) != 1 {
		t.Fatalf("esperaba exactamente la clave %q marcada, obtuve %+v", clave, leidas)
	}

	// Idempotente: volver a marcarla no falla ni duplica la fila.
	if err := repo.MarcarLeida(clave, domain.NotifContratoPorVencer, domain.EntidadContrato, 42); err != nil {
		t.Fatalf("MarcarLeida repetida: %v", err)
	}
	leidas, _ = repo.ClavesLeidas()
	if len(leidas) != 1 {
		t.Fatalf("marcar dos veces la misma clave no debería crear dos filas, hay %d", len(leidas))
	}
}
