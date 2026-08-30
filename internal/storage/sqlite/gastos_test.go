package sqlite_test

import (
	"testing"
	"time"

	"github.com/naparito/Arantxator-flat-admin/internal/domain"
	"github.com/naparito/Arantxator-flat-admin/internal/storage/sqlite"
)

func gastoBase(inmuebleID int64) domain.Gasto {
	venc := fecha(2026, 9, 30)
	return domain.Gasto{
		InmuebleID:       inmuebleID,
		Tipo:             domain.GastoLuz,
		Periodicidad:     domain.PeriodicidadMensual,
		Importe:          78.00,
		FechaEmision:     fecha(2026, 9, 1),
		FechaVencimiento: &venc,
		Proveedor:        "Iberdrola",
	}
}

func TestGastosRepo_CrearLeerYListar(t *testing.T) {
	db := newTestDB(t)
	inmuebles := sqlite.NewInmueblesRepo(db)
	gastos := sqlite.NewGastosRepo(db)
	inmueble := crearInmuebleSimple(t, inmuebles)

	creado, err := gastos.Create(gastoBase(inmueble.ID))
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if creado.ID == 0 || creado.Importe != 78.00 || creado.Tipo != domain.GastoLuz {
		t.Fatalf("gasto creado inesperado: %+v", creado)
	}
	if creado.EstadoPago != domain.PagoPendiente {
		t.Fatalf("un gasto nuevo debería nacer 'pendiente', obtuve %q", creado.EstadoPago)
	}
	// Las fechas DATE viajan como AAAA-MM-DD y se releen igual.
	if time.Time(creado.FechaEmision).Format("2006-01-02") != "2026-09-01" {
		t.Fatalf("fecha de emisión inesperada: %v", time.Time(creado.FechaEmision))
	}

	leido, err := gastos.Get(creado.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if leido.Proveedor != "Iberdrola" || leido.Periodicidad != domain.PeriodicidadMensual {
		t.Fatalf("gasto releído inesperado: %+v", leido)
	}

	lista, err := gastos.ListByInmueble(inmueble.ID)
	if err != nil {
		t.Fatalf("ListByInmueble: %v", err)
	}
	if len(lista) != 1 || lista[0].ID != creado.ID {
		t.Fatalf("esperaba 1 gasto en el listado del inmueble, obtuve %+v", lista)
	}
}

func TestGastosRepo_MarcarPagadaSellaLaFechaDePago(t *testing.T) {
	db := newTestDB(t)
	inmuebles := sqlite.NewInmueblesRepo(db)
	gastos := sqlite.NewGastosRepo(db)
	inmueble := crearInmuebleSimple(t, inmuebles)

	creado, err := gastos.Create(gastoBase(inmueble.ID))
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	g := creado
	g.EstadoPago = domain.PagoPagado
	g.MetodoPago = "transferencia"
	pagado, err := gastos.Update(creado.ID, g)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if pagado.EstadoPago != domain.PagoPagado {
		t.Fatalf("esperaba estado 'pagado', obtuve %q", pagado.EstadoPago)
	}
	if pagado.FechaPago == nil || time.Time(*pagado.FechaPago).IsZero() {
		t.Fatal("al marcar pagada sin fecha explícita se debería sellar con hoy")
	}

	// Volver a pendiente limpia la fecha de pago.
	g2 := pagado
	g2.EstadoPago = domain.PagoPendiente
	vuelta, err := gastos.Update(creado.ID, g2)
	if err != nil {
		t.Fatalf("Update de vuelta: %v", err)
	}
	if vuelta.FechaPago != nil {
		t.Fatalf("al volver a pendiente la fecha de pago debería limpiarse, obtuve %v", vuelta.FechaPago)
	}
}

func TestGastosRepo_BorrarInmuebleArrastraLosGastos(t *testing.T) {
	db := newTestDB(t)
	inmuebles := sqlite.NewInmueblesRepo(db)
	gastos := sqlite.NewGastosRepo(db)
	inmueble := crearInmuebleSimple(t, inmuebles)

	creado, err := gastos.Create(gastoBase(inmueble.ID))
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, err := db.Exec(`DELETE FROM inmuebles WHERE id = ?`, inmueble.ID); err != nil {
		t.Fatalf("DELETE inmueble: %v", err)
	}
	if _, err := gastos.Get(creado.ID); err != sqlite.ErrNotFound {
		t.Fatalf("el gasto debería borrarse en cascada con su inmueble, err = %v", err)
	}
}

func TestGastosRepo_SumaImporteEnPeriodo(t *testing.T) {
	db := newTestDB(t)
	inmuebles := sqlite.NewInmueblesRepo(db)
	gastos := sqlite.NewGastosRepo(db)
	inmueble := crearInmuebleSimple(t, inmuebles)

	dentro1 := gastoBase(inmueble.ID)
	dentro1.FechaEmision = fecha(2026, 9, 1)
	dentro1.Importe = 78
	dentro2 := gastoBase(inmueble.ID)
	dentro2.Tipo = domain.GastoAgua
	dentro2.FechaEmision = fecha(2026, 9, 28)
	dentro2.Importe = 42
	fuera := gastoBase(inmueble.ID)
	fuera.Tipo = domain.GastoGas
	fuera.FechaEmision = fecha(2026, 10, 2)
	fuera.Importe = 35
	for _, g := range []domain.Gasto{dentro1, dentro2, fuera} {
		if _, err := gastos.Create(g); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	desde := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	hasta := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	total, err := gastos.SumaImporteEnPeriodo(inmueble.ID, desde, hasta)
	if err != nil {
		t.Fatalf("SumaImporteEnPeriodo: %v", err)
	}
	if total != 120 {
		t.Fatalf("esperaba 120 € (78 + 42, sin el de octubre), obtuve %v", total)
	}
}
