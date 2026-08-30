package sqlite_test

import (
	"testing"
	"time"

	"github.com/naparito/Arantxator-flat-admin/internal/domain"
	"github.com/naparito/Arantxator-flat-admin/internal/storage/sqlite"
)

func TestRepartosRepo_CrearVersionYLeerLaVigente(t *testing.T) {
	db := newTestDB(t)
	inmuebles := sqlite.NewInmueblesRepo(db)
	inquilinos := sqlite.NewInquilinosRepo(db)
	repartos := sqlite.NewRepartosRepo(db)

	inmueble := crearInmuebleCompartido(t, inmuebles)
	a := crearInquilinoRapido(t, inquilinos, "Ana")
	b := crearInquilinoRapido(t, inquilinos, "Bruno")
	c := crearInquilinoRapido(t, inquilinos, "Carla")

	cuotas := []sqlite.Cuota{
		{InquilinoID: a.ID, Porcentaje: 33},
		{InquilinoID: b.ID, Porcentaje: 33},
		{InquilinoID: c.ID, Porcentaje: 34},
	}
	vigente, err := repartos.CrearVersion(inmueble.ID, domain.GastoLuz, fecha(2026, 3, 1), "entrada de Carla", cuotas)
	if err != nil {
		t.Fatalf("CrearVersion: %v", err)
	}
	if len(vigente) != 3 {
		t.Fatalf("esperaba 3 cuotas en la versión vigente, obtuve %+v", vigente)
	}
	if vigente[0].Motivo != "entrada de Carla" {
		t.Fatalf("esperaba el motivo guardado, obtuve %q", vigente[0].Motivo)
	}

	leida, err := repartos.VigenteEnFecha(inmueble.ID, domain.GastoLuz, time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("VigenteEnFecha: %v", err)
	}
	suma := 0.0
	for _, r := range leida {
		suma += r.Porcentaje
	}
	if suma != 100 {
		t.Fatalf("los porcentajes de la versión vigente deberían sumar 100, suman %v", suma)
	}
}

func TestRepartosRepo_NuevaVersionCierraLaAnteriorSinBorrarla(t *testing.T) {
	db := newTestDB(t)
	inmuebles := sqlite.NewInmueblesRepo(db)
	inquilinos := sqlite.NewInquilinosRepo(db)
	repartos := sqlite.NewRepartosRepo(db)

	inmueble := crearInmuebleCompartido(t, inmuebles)
	a := crearInquilinoRapido(t, inquilinos, "Ana")
	b := crearInquilinoRapido(t, inquilinos, "Bruno")
	c := crearInquilinoRapido(t, inquilinos, "Carla")

	// v1: dos inquilinos al 50 %, vigente desde 2026-01-01.
	if _, err := repartos.CrearVersion(inmueble.ID, domain.GastoLuz, fecha(2026, 1, 1), "reparto inicial", []sqlite.Cuota{
		{InquilinoID: a.ID, Porcentaje: 50},
		{InquilinoID: b.ID, Porcentaje: 50},
	}); err != nil {
		t.Fatalf("CrearVersion v1: %v", err)
	}

	// v2: entra Carla, 33/33/34, vigente desde 2026-03-01.
	if _, err := repartos.CrearVersion(inmueble.ID, domain.GastoLuz, fecha(2026, 3, 1), "entrada de Carla", []sqlite.Cuota{
		{InquilinoID: a.ID, Porcentaje: 33},
		{InquilinoID: b.ID, Porcentaje: 33},
		{InquilinoID: c.ID, Porcentaje: 34},
	}); err != nil {
		t.Fatalf("CrearVersion v2: %v", err)
	}

	todas, err := repartos.ListByInmueble(inmueble.ID)
	if err != nil {
		t.Fatalf("ListByInmueble: %v", err)
	}
	if len(todas) != 5 {
		t.Fatalf("esperaba 5 filas (2 de v1 + 3 de v2), la v1 no se borra: obtuve %d", len(todas))
	}
	// La v1 quedó cerrada con vigente_hasta = 2026-03-01.
	cerradas := 0
	for _, r := range todas {
		if r.VigenteHasta != nil && time.Time(*r.VigenteHasta).Format("2006-01-02") == "2026-03-01" {
			cerradas++
		}
	}
	if cerradas != 2 {
		t.Fatalf("esperaba 2 filas de la v1 cerradas al 2026-03-01, obtuve %d", cerradas)
	}

	// Una factura de febrero usa el reparto viejo (50/50, 2 inquilinos).
	feb, err := repartos.VigenteEnFecha(inmueble.ID, domain.GastoLuz, time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("VigenteEnFecha feb: %v", err)
	}
	if len(feb) != 2 || feb[0].Porcentaje != 50 {
		t.Fatalf("una factura de febrero debería usar la v1 (2 al 50%%), obtuve %+v", feb)
	}
	// Una de septiembre usa la nueva (33/33/34, 3 inquilinos).
	sep, err := repartos.VigenteEnFecha(inmueble.ID, domain.GastoLuz, time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("VigenteEnFecha sep: %v", err)
	}
	if len(sep) != 3 {
		t.Fatalf("una factura de septiembre debería usar la v2 (3 inquilinos), obtuve %+v", sep)
	}
}
