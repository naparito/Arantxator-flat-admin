package sqlite_test

import (
	"testing"
	"time"

	"github.com/naparito/Arantxator-flat-admin/internal/domain"
	"github.com/naparito/Arantxator-flat-admin/internal/storage/sqlite"
)

func fecha(y int, m time.Month, d int) domain.Fecha {
	return domain.Fecha(time.Date(y, m, d, 0, 0, 0, 0, time.UTC))
}

func crearInmuebleSimple(t *testing.T, inmuebles *sqlite.InmueblesRepo) domain.Inmueble {
	t.Helper()
	creado, err := inmuebles.Create(domain.Inmueble{Nombre: "Alcalá 145", Direccion: "Calle de Alcalá 145, 3ºB", Tipo: domain.TipoPiso})
	if err != nil {
		t.Fatalf("Create inmueble: %v", err)
	}
	return creado
}

func crearInquilinoRapido(t *testing.T, inquilinos *sqlite.InquilinosRepo, nombre string) domain.Inquilino {
	t.Helper()
	creado, err := inquilinos.Create(domain.Inquilino{NombreCompleto: nombre, DocumentoIdentidad: nombre})
	if err != nil {
		t.Fatalf("Create inquilino %s: %v", nombre, err)
	}
	return creado
}

func contratoBase(inmuebleID int64, inquilinoIDs ...int64) domain.Contrato {
	return domain.Contrato{
		InmuebleID:   inmuebleID,
		InquilinoIDs: inquilinoIDs,
		FechaFirma:   fecha(2026, 2, 1),
		FechaInicio:  fecha(2026, 2, 1),
		FechaFin:     fecha(2031, 1, 31),
		RentaMensual: 980,
		DiaPago:      5,
	}
}

func TestContratosRepo_CreateConVariosInquilinos(t *testing.T) {
	db := newTestDB(t)
	inmuebles := sqlite.NewInmueblesRepo(db)
	inquilinos := sqlite.NewInquilinosRepo(db)
	contratos := sqlite.NewContratosRepo(db)

	inmueble := crearInmuebleSimple(t, inmuebles)
	a := crearInquilinoRapido(t, inquilinos, "Ana")
	b := crearInquilinoRapido(t, inquilinos, "Bruno")
	c := crearInquilinoRapido(t, inquilinos, "Carla")

	creado, err := contratos.Create(contratoBase(inmueble.ID, a.ID, b.ID, c.ID))
	if err != nil {
		t.Fatalf("Create contrato: %v", err)
	}
	if len(creado.InquilinoIDs) != 3 {
		t.Fatalf("esperaba 3 co-arrendatarios vinculados, obtuve %v", creado.InquilinoIDs)
	}

	leido, err := contratos.Get(creado.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if len(leido.InquilinoIDs) != 3 {
		t.Fatalf("Get: esperaba 3 co-arrendatarios, obtuve %v", leido.InquilinoIDs)
	}
	if time.Time(leido.FechaFirma).IsZero() || time.Time(leido.FechaFin).IsZero() {
		t.Fatalf("las fechas del contrato no se persistieron: %+v", leido)
	}

	// El histórico de cada inquilino recoge este contrato.
	for _, id := range []int64{a.ID, b.ID, c.ID} {
		hist, err := contratos.ListByInquilino(id)
		if err != nil {
			t.Fatalf("ListByInquilino(%d): %v", id, err)
		}
		if len(hist) != 1 || hist[0].ID != creado.ID {
			t.Fatalf("inquilino %d: esperaba el contrato %d en su histórico, obtuve %+v", id, creado.ID, hist)
		}
	}
}

func TestContratosRepo_InmuebleNoCompartidoSeAlquilaYSeLibera(t *testing.T) {
	db := newTestDB(t)
	inmuebles := sqlite.NewInmueblesRepo(db)
	inquilinos := sqlite.NewInquilinosRepo(db)
	contratos := sqlite.NewContratosRepo(db)

	inmueble := crearInmuebleSimple(t, inmuebles)
	inq := crearInquilinoRapido(t, inquilinos, "Ana")

	entrada := contratoBase(inmueble.ID, inq.ID)
	entrada.FechaFin = fecha(2030, 1, 31)
	creado, err := contratos.Create(entrada)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	trasAlta, _ := inmuebles.Get(inmueble.ID)
	if trasAlta.Estado != domain.InmuebleAlquilado {
		t.Fatalf("esperaba el inmueble en 'alquilado' tras activar el contrato, obtuve %q", trasAlta.Estado)
	}

	rescindido := creado
	rescindido.Estado = domain.ContratoRescindido
	rescindido.MotivoBaja = "acuerdo entre las partes"
	if _, err := contratos.Update(creado.ID, rescindido); err != nil {
		t.Fatalf("Update rescindir: %v", err)
	}

	trasBaja, _ := inmuebles.Get(inmueble.ID)
	if trasBaja.Estado != domain.InmuebleDisponible {
		t.Fatalf("esperaba el inmueble de vuelta en 'disponible' tras rescindir, obtuve %q", trasBaja.Estado)
	}
}

func TestContratosRepo_InmuebleCompartidoNoCambiaEstado(t *testing.T) {
	db := newTestDB(t)
	inmuebles := sqlite.NewInmueblesRepo(db)
	inquilinos := sqlite.NewInquilinosRepo(db)
	habitaciones := sqlite.NewHabitacionesRepo(db)
	contratos := sqlite.NewContratosRepo(db)

	inmueble := crearInmuebleCompartido(t, inmuebles)
	hab, err := habitaciones.Create(domain.Habitacion{InmuebleID: inmueble.ID, Nombre: "Habitación 1"})
	if err != nil {
		t.Fatalf("Create habitación: %v", err)
	}
	inq := crearInquilinoRapido(t, inquilinos, "Ana")

	entrada := contratoBase(inmueble.ID, inq.ID)
	entrada.HabitacionID = &hab.ID
	entrada.FechaFin = fecha(2030, 1, 31)
	if _, err := contratos.Create(entrada); err != nil {
		t.Fatalf("Create: %v", err)
	}

	tras, _ := inmuebles.Get(inmueble.ID)
	if tras.Estado != domain.InmuebleDisponible {
		t.Fatalf("un inmueble compartido no debe cambiar de estado al activar un contrato de habitación, obtuve %q", tras.Estado)
	}
}

func TestContratosRepo_NoSolapamientoPorHabitacion(t *testing.T) {
	db := newTestDB(t)
	inmuebles := sqlite.NewInmueblesRepo(db)
	inquilinos := sqlite.NewInquilinosRepo(db)
	habitaciones := sqlite.NewHabitacionesRepo(db)
	contratos := sqlite.NewContratosRepo(db)

	inmueble := crearInmuebleCompartido(t, inmuebles)
	hab1, _ := habitaciones.Create(domain.Habitacion{InmuebleID: inmueble.ID, Nombre: "Habitación 1"})
	hab2, _ := habitaciones.Create(domain.Habitacion{InmuebleID: inmueble.ID, Nombre: "Habitación 2"})
	inq := crearInquilinoRapido(t, inquilinos, "Ana")

	c1 := contratoBase(inmueble.ID, inq.ID)
	c1.HabitacionID = &hab1.ID
	c1.FechaFin = fecha(2030, 1, 31)
	if _, err := contratos.Create(c1); err != nil {
		t.Fatalf("Create contrato hab1: %v", err)
	}

	ref := time.Now()
	ocupadaHab1, err := contratos.TieneContratoVigenteEnAmbito(inmueble.ID, &hab1.ID, 0, ref)
	if err != nil {
		t.Fatalf("TieneContratoVigenteEnAmbito hab1: %v", err)
	}
	if !ocupadaHab1 {
		t.Fatal("la habitación 1 debería contar como ocupada por un contrato vigente")
	}
	libreHab2, err := contratos.TieneContratoVigenteEnAmbito(inmueble.ID, &hab2.ID, 0, ref)
	if err != nil {
		t.Fatalf("TieneContratoVigenteEnAmbito hab2: %v", err)
	}
	if libreHab2 {
		t.Fatal("la habitación 2 no tiene contratos: no debería contar como ocupada")
	}
}

func TestContratosRepo_NoSolapamientoPorInmueble(t *testing.T) {
	db := newTestDB(t)
	inmuebles := sqlite.NewInmueblesRepo(db)
	inquilinos := sqlite.NewInquilinosRepo(db)
	contratos := sqlite.NewContratosRepo(db)

	inmueble := crearInmuebleSimple(t, inmuebles)
	inq := crearInquilinoRapido(t, inquilinos, "Ana")

	c := contratoBase(inmueble.ID, inq.ID)
	c.FechaFin = fecha(2030, 1, 31)
	if _, err := contratos.Create(c); err != nil {
		t.Fatalf("Create: %v", err)
	}

	ocupado, err := contratos.TieneContratoVigenteEnAmbito(inmueble.ID, nil, 0, time.Now())
	if err != nil {
		t.Fatalf("TieneContratoVigenteEnAmbito: %v", err)
	}
	if !ocupado {
		t.Fatal("un inmueble no compartido con un contrato vigente debería contar como ocupado")
	}
}

func TestContratosRepo_OcupacionInmueble(t *testing.T) {
	db := newTestDB(t)
	inmuebles := sqlite.NewInmueblesRepo(db)
	inquilinos := sqlite.NewInquilinosRepo(db)
	habitaciones := sqlite.NewHabitacionesRepo(db)
	contratos := sqlite.NewContratosRepo(db)

	inmueble := crearInmuebleCompartido(t, inmuebles)
	hab1, _ := habitaciones.Create(domain.Habitacion{InmuebleID: inmueble.ID, Nombre: "Habitación 1"})
	hab2, _ := habitaciones.Create(domain.Habitacion{InmuebleID: inmueble.ID, Nombre: "Habitación 2"})
	habitaciones.Create(domain.Habitacion{InmuebleID: inmueble.ID, Nombre: "Habitación 3"})

	nombres := map[int64]string{hab1.ID: "Ana", hab2.ID: "Bruno"}
	for _, habID := range []int64{hab1.ID, hab2.ID} {
		inq := crearInquilinoRapido(t, inquilinos, nombres[habID])
		c := contratoBase(inmueble.ID, inq.ID)
		id := habID
		c.HabitacionID = &id
		c.FechaFin = fecha(2030, 1, 31)
		if _, err := contratos.Create(c); err != nil {
			t.Fatalf("Create contrato hab %d: %v", habID, err)
		}
	}

	ocupadas, totales, err := contratos.OcupacionInmueble(inmueble.ID, time.Now())
	if err != nil {
		t.Fatalf("OcupacionInmueble: %v", err)
	}
	if ocupadas != 2 || totales != 3 {
		t.Fatalf("esperaba 2/3 habitaciones ocupadas, obtuve %d/%d", ocupadas, totales)
	}
}

func TestContratosRepo_BorrarInquilinoConContratoSeBloquea(t *testing.T) {
	db := newTestDB(t)
	inmuebles := sqlite.NewInmueblesRepo(db)
	inquilinos := sqlite.NewInquilinosRepo(db)
	contratos := sqlite.NewContratosRepo(db)

	inmueble := crearInmuebleSimple(t, inmuebles)
	inq := crearInquilinoRapido(t, inquilinos, "Ana")
	if _, err := contratos.Create(contratoBase(inmueble.ID, inq.ID)); err != nil {
		t.Fatalf("Create: %v", err)
	}

	if _, err := db.Exec(`DELETE FROM inquilinos WHERE id = ?`, inq.ID); err == nil {
		t.Fatal("esperaba que ON DELETE RESTRICT bloqueara el borrado de un inquilino con contrato, pero no hubo error")
	}
}

func TestContratosRepo_BorrarHabitacionConContratoSeBloquea(t *testing.T) {
	db := newTestDB(t)
	inmuebles := sqlite.NewInmueblesRepo(db)
	inquilinos := sqlite.NewInquilinosRepo(db)
	habitaciones := sqlite.NewHabitacionesRepo(db)
	contratos := sqlite.NewContratosRepo(db)

	inmueble := crearInmuebleCompartido(t, inmuebles)
	hab, _ := habitaciones.Create(domain.Habitacion{InmuebleID: inmueble.ID, Nombre: "Habitación 1"})
	inq := crearInquilinoRapido(t, inquilinos, "Ana")

	c := contratoBase(inmueble.ID, inq.ID)
	c.HabitacionID = &hab.ID
	if _, err := contratos.Create(c); err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := habitaciones.Delete(hab.ID); err == nil {
		t.Fatal("esperaba que ON DELETE RESTRICT bloqueara el borrado de una habitación con contrato, pero no hubo error")
	}
}
