package sqlite_test

import (
	"testing"
	"time"

	"github.com/naparito/Arantxator-flat-admin/internal/domain"
	"github.com/naparito/Arantxator-flat-admin/internal/storage/sqlite"
)

func TestCobrosRepo_CrearListarYSumarPeriodo(t *testing.T) {
	db := newTestDB(t)
	inmuebles := sqlite.NewInmueblesRepo(db)
	cobros := sqlite.NewCobrosRepo(db)
	inmueble := crearInmuebleSimple(t, inmuebles)

	cobroSep := domain.CobroRenta{
		InmuebleID: inmueble.ID,
		Periodo:    fecha(2026, 9, 1),
		Importe:    1350,
		MetodoPago: "transferencia",
	}
	cobroOct := domain.CobroRenta{
		InmuebleID: inmueble.ID,
		Periodo:    fecha(2026, 10, 1),
		Importe:    1350,
	}
	for _, c := range []domain.CobroRenta{cobroSep, cobroOct} {
		if _, err := cobros.Create(c); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	lista, err := cobros.ListByInmueble(inmueble.ID)
	if err != nil {
		t.Fatalf("ListByInmueble: %v", err)
	}
	if len(lista) != 2 {
		t.Fatalf("esperaba 2 cobros, obtuve %d", len(lista))
	}

	desde := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	hasta := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	total, err := cobros.SumaEnPeriodo(inmueble.ID, desde, hasta)
	if err != nil {
		t.Fatalf("SumaEnPeriodo: %v", err)
	}
	if total != 1350 {
		t.Fatalf("esperaba 1350 € (solo el de septiembre), obtuve %v", total)
	}
}

func TestCobrosRepo_BorrarInmuebleArrastraLosCobros(t *testing.T) {
	db := newTestDB(t)
	inmuebles := sqlite.NewInmueblesRepo(db)
	cobros := sqlite.NewCobrosRepo(db)
	inmueble := crearInmuebleSimple(t, inmuebles)

	creado, err := cobros.Create(domain.CobroRenta{InmuebleID: inmueble.ID, Periodo: fecha(2026, 9, 1), Importe: 1350})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, err := db.Exec(`DELETE FROM inmuebles WHERE id = ?`, inmueble.ID); err != nil {
		t.Fatalf("DELETE inmueble: %v", err)
	}
	if _, err := cobros.Get(creado.ID); err != sqlite.ErrNotFound {
		t.Fatalf("el cobro debería borrarse en cascada con su inmueble, err = %v", err)
	}
}
