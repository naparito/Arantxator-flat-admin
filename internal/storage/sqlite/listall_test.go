package sqlite_test

import (
	"testing"

	"github.com/naparito/Arantxator-flat-admin/internal/domain"
	"github.com/naparito/Arantxator-flat-admin/internal/storage/sqlite"
)

// Los List() sin filtro de gastos, incidencias y cobros alimentan el resumen
// del dashboard y el motor de reglas de notificación (Hito 6): tienen que
// devolver filas de TODA la cartera, no de un inmueble concreto.
func TestListAllCruzaTodaLaCartera(t *testing.T) {
	db := newTestDB(t)
	inmuebles := sqlite.NewInmueblesRepo(db)
	gastos := sqlite.NewGastosRepo(db)
	incidencias := sqlite.NewIncidenciasRepo(db)
	cobros := sqlite.NewCobrosRepo(db)

	a, err := inmuebles.Create(domain.Inmueble{Nombre: "A", Direccion: "Calle A 1", Tipo: domain.TipoPiso})
	if err != nil {
		t.Fatalf("Create inmueble A: %v", err)
	}
	b, err := inmuebles.Create(domain.Inmueble{Nombre: "B", Direccion: "Calle B 2", Tipo: domain.TipoPiso})
	if err != nil {
		t.Fatalf("Create inmueble B: %v", err)
	}

	if _, err := gastos.Create(gastoBase(a.ID)); err != nil {
		t.Fatalf("Create gasto A: %v", err)
	}
	if _, err := gastos.Create(gastoBase(b.ID)); err != nil {
		t.Fatalf("Create gasto B: %v", err)
	}
	if _, err := incidencias.Create(incidenciaBase(a.ID)); err != nil {
		t.Fatalf("Create incidencia A: %v", err)
	}
	if _, err := incidencias.Create(incidenciaBase(b.ID)); err != nil {
		t.Fatalf("Create incidencia B: %v", err)
	}
	for _, m := range []domain.Inmueble{a, b} {
		if _, err := cobros.Create(domain.CobroRenta{InmuebleID: m.ID, Periodo: fecha(2026, 8, 1), Importe: 900}); err != nil {
			t.Fatalf("Create cobro %s: %v", m.Nombre, err)
		}
	}

	if gs, err := gastos.List(); err != nil || len(gs) != 2 {
		t.Fatalf("gastos.List(): esperaba 2 filas de toda la cartera, obtuve %d (err %v)", len(gs), err)
	}
	if is, err := incidencias.List(); err != nil || len(is) != 2 {
		t.Fatalf("incidencias.List(): esperaba 2 filas, obtuve %d (err %v)", len(is), err)
	}
	if cs, err := cobros.List(); err != nil || len(cs) != 2 {
		t.Fatalf("cobros.List(): esperaba 2 filas, obtuve %d (err %v)", len(cs), err)
	}
}
