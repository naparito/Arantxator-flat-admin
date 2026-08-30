package sqlite_test

import (
	"testing"

	"github.com/naparito/Arantxator-flat-admin/internal/domain"
	"github.com/naparito/Arantxator-flat-admin/internal/storage/sqlite"
)

func crearInmuebleCompartido(t *testing.T, inmuebles *sqlite.InmueblesRepo) domain.Inmueble {
	t.Helper()
	creado, err := inmuebles.Create(domain.Inmueble{
		Nombre: "Bravo Murillo 210", Direccion: "dir", Tipo: domain.TipoPiso, Compartido: true,
	})
	if err != nil {
		t.Fatalf("Create inmueble compartido: %v", err)
	}
	if !creado.Compartido {
		t.Fatalf("esperaba compartido=true, obtuve %+v", creado)
	}
	return creado
}

func TestInmueblesRepo_Compartido(t *testing.T) {
	db := newTestDB(t)
	inmuebles := sqlite.NewInmueblesRepo(db)

	crearInmuebleCompartido(t, inmuebles)

	noCompartido, err := inmuebles.Create(domain.Inmueble{Nombre: "x", Direccion: "y", Tipo: domain.TipoPiso})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if noCompartido.Compartido {
		t.Fatalf("esperaba compartido=false por defecto, obtuve true")
	}
}

func TestHabitacionesRepo_CrearYListar(t *testing.T) {
	db := newTestDB(t)
	inmuebles := sqlite.NewInmueblesRepo(db)
	habitaciones := sqlite.NewHabitacionesRepo(db)

	inmueble := crearInmuebleCompartido(t, inmuebles)

	h1, err := habitaciones.Create(domain.Habitacion{InmuebleID: inmueble.ID, Nombre: "Habitación 1", M2: 12, TieneBano: true})
	if err != nil {
		t.Fatalf("Create habitación 1: %v", err)
	}
	if h1.ID == 0 || h1.InmuebleID != inmueble.ID || !h1.TieneBano {
		t.Fatalf("habitación creada inesperada: %+v", h1)
	}
	if h1.InquilinoID != nil {
		t.Fatalf("esperaba una habitación sin ocupante recién creada, obtuve %+v", h1.InquilinoID)
	}

	if _, err := habitaciones.Create(domain.Habitacion{InmuebleID: inmueble.ID, Nombre: "Habitación 2", M2: 10}); err != nil {
		t.Fatalf("Create habitación 2: %v", err)
	}

	lista, err := habitaciones.ListByInmueble(inmueble.ID)
	if err != nil {
		t.Fatalf("ListByInmueble: %v", err)
	}
	if len(lista) != 2 {
		t.Fatalf("esperaba 2 habitaciones, obtuve %d: %+v", len(lista), lista)
	}
}

func TestHabitacionesRepo_ListarInmuebleInexistenteDevuelveVacio(t *testing.T) {
	db := newTestDB(t)
	habitaciones := sqlite.NewHabitacionesRepo(db)

	lista, err := habitaciones.ListByInmueble(999)
	if err != nil {
		t.Fatalf("ListByInmueble: %v", err)
	}
	if len(lista) != 0 {
		t.Fatalf("esperaba lista vacía, obtuve %+v", lista)
	}
}

func TestHabitacionesRepo_Update(t *testing.T) {
	db := newTestDB(t)
	inmuebles := sqlite.NewInmueblesRepo(db)
	habitaciones := sqlite.NewHabitacionesRepo(db)

	inmueble := crearInmuebleCompartido(t, inmuebles)
	creada, err := habitaciones.Create(domain.Habitacion{InmuebleID: inmueble.ID, Nombre: "Habitación 1"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	actualizada, err := habitaciones.Update(creada.ID, domain.Habitacion{Nombre: "Doble exterior", M2: 15, TieneBano: true, Amueblada: true, Notas: "con balcón"})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if actualizada.Nombre != "Doble exterior" || actualizada.M2 != 15 || !actualizada.TieneBano || !actualizada.Amueblada || actualizada.Notas != "con balcón" {
		t.Fatalf("los cambios no se aplicaron: %+v", actualizada)
	}
}

func TestHabitacionesRepo_UpdateIDInexistente(t *testing.T) {
	habitaciones := sqlite.NewHabitacionesRepo(newTestDB(t))
	_, err := habitaciones.Update(999, domain.Habitacion{Nombre: "x"})
	if err != sqlite.ErrNotFound {
		t.Fatalf("esperaba ErrNotFound, obtuve %v", err)
	}
}

func TestHabitacionesRepo_AsignarYQuitarOcupante(t *testing.T) {
	db := newTestDB(t)
	inmuebles := sqlite.NewInmueblesRepo(db)
	habitaciones := sqlite.NewHabitacionesRepo(db)

	inmueble := crearInmuebleCompartido(t, inmuebles)
	creada, err := habitaciones.Create(domain.Habitacion{InmuebleID: inmueble.ID, Nombre: "Habitación 1"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	inquilinoID := int64(1) // la tabla inquilinos existe desde el Hito 0; su alta llega en el Hito 2
	if _, err := db.Exec(`INSERT INTO inquilinos (id, nombre_completo, documento_identidad) VALUES (?, 'x', 'y')`, inquilinoID); err != nil {
		t.Fatalf("preparando inquilino de prueba: %v", err)
	}

	asignada, err := habitaciones.AsignarOcupante(creada.ID, &inquilinoID)
	if err != nil {
		t.Fatalf("AsignarOcupante: %v", err)
	}
	if asignada.InquilinoID == nil || *asignada.InquilinoID != inquilinoID {
		t.Fatalf("esperaba inquilino_id=%d, obtuve %+v", inquilinoID, asignada.InquilinoID)
	}

	liberada, err := habitaciones.AsignarOcupante(creada.ID, nil)
	if err != nil {
		t.Fatalf("AsignarOcupante(nil): %v", err)
	}
	if liberada.InquilinoID != nil {
		t.Fatalf("esperaba habitación libre, obtuve ocupante %+v", liberada.InquilinoID)
	}
}

func TestHabitacionesRepo_Delete(t *testing.T) {
	db := newTestDB(t)
	inmuebles := sqlite.NewInmueblesRepo(db)
	habitaciones := sqlite.NewHabitacionesRepo(db)

	inmueble := crearInmuebleCompartido(t, inmuebles)
	creada, err := habitaciones.Create(domain.Habitacion{InmuebleID: inmueble.ID, Nombre: "Habitación 1"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := habitaciones.Delete(creada.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := habitaciones.Get(creada.ID); err != sqlite.ErrNotFound {
		t.Fatalf("esperaba ErrNotFound tras borrar, obtuve %v", err)
	}
}

func TestHabitacionesRepo_SeBorranEnCascadaConElInmueble(t *testing.T) {
	db := newTestDB(t)
	inmuebles := sqlite.NewInmueblesRepo(db)
	habitaciones := sqlite.NewHabitacionesRepo(db)

	inmueble := crearInmuebleCompartido(t, inmuebles)
	creada, err := habitaciones.Create(domain.Habitacion{InmuebleID: inmueble.ID, Nombre: "Habitación 1"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if _, err := db.Exec(`DELETE FROM inmuebles WHERE id = ?`, inmueble.ID); err != nil {
		t.Fatalf("borrando inmueble: %v", err)
	}

	if _, err := habitaciones.Get(creada.ID); err != sqlite.ErrNotFound {
		t.Fatalf("esperaba que la habitación se borrara en cascada, obtuve err=%v", err)
	}
}
