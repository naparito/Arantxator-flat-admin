package sqlite_test

import (
	"testing"
	"time"

	"github.com/naparito/Arantxator-flat-admin/internal/domain"
	"github.com/naparito/Arantxator-flat-admin/internal/storage/sqlite"
)

func TestInquilinosRepo_CreateConTodosLosCampos(t *testing.T) {
	repo := sqlite.NewInquilinosRepo(newTestDB(t))
	nacimiento := time.Date(1992, 3, 14, 0, 0, 0, 0, time.UTC)

	entrada := domain.Inquilino{
		NombreCompleto:             "Laura Fernández Ruiz",
		DocumentoIdentidad:         "45123456M",
		FechaNacimiento:            &nacimiento,
		Telefono:                   "+34 611 223 344",
		Email:                      "laura.fr@email.com",
		Nacionalidad:               "Española",
		ContactoEmergenciaNombre:   "Marisol Ruiz Peña (madre)",
		ContactoEmergenciaTelefono: "+34 622 990 011",
		IBAN:                       "ES9121000000000000001234",
	}

	creado, err := repo.Create(entrada)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if creado.ID == 0 {
		t.Fatalf("esperaba un id asignado, obtuve 0")
	}

	leido, err := repo.Get(creado.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if leido.NombreCompleto != entrada.NombreCompleto ||
		leido.DocumentoIdentidad != entrada.DocumentoIdentidad ||
		leido.Telefono != entrada.Telefono ||
		leido.Email != entrada.Email ||
		leido.Nacionalidad != entrada.Nacionalidad ||
		leido.ContactoEmergenciaNombre != entrada.ContactoEmergenciaNombre ||
		leido.ContactoEmergenciaTelefono != entrada.ContactoEmergenciaTelefono ||
		leido.IBAN != entrada.IBAN {
		t.Fatalf("el inquilino leído no coincide con el creado:\n  creado=%+v\n  leído =%+v", entrada, leido)
	}
	if leido.FechaNacimiento == nil || !leido.FechaNacimiento.Equal(nacimiento) {
		t.Fatalf("fecha de nacimiento no coincide: %v", leido.FechaNacimiento)
	}
	if leido.CreadoEn.IsZero() || leido.ActualizadoEn.IsZero() {
		t.Fatalf("esperaba creado_en/actualizado_en no vacíos")
	}
}

func TestInquilinosRepo_CreateSoloConCamposObligatorios(t *testing.T) {
	repo := sqlite.NewInquilinosRepo(newTestDB(t))

	creado, err := repo.Create(domain.Inquilino{
		NombreCompleto:     "Ana Belén Torres",
		DocumentoIdentidad: "50987654K",
	})
	if err != nil {
		t.Fatalf("Create con campos mínimos: %v", err)
	}

	leido, err := repo.Get(creado.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if leido.Telefono != "" || leido.Email != "" || leido.Nacionalidad != "" || leido.IBAN != "" {
		t.Fatalf("esperaba campos opcionales vacíos, obtuve %+v", leido)
	}
	if leido.FechaNacimiento != nil {
		t.Fatalf("esperaba fecha de nacimiento nula, obtuve %v", leido.FechaNacimiento)
	}
}

func TestInquilinosRepo_Update(t *testing.T) {
	repo := sqlite.NewInquilinosRepo(newTestDB(t))

	creado, err := repo.Create(domain.Inquilino{NombreCompleto: "Javier Martín Soto", DocumentoIdentidad: "Y1234567L"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	time.Sleep(1100 * time.Millisecond) // CURRENT_TIMESTAMP de SQLite tiene resolución de segundo

	editado := creado
	editado.Telefono = "+34 600 111 222"
	editado.Email = "javier.ms@email.com"

	actualizado, err := repo.Update(creado.ID, editado)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if actualizado.Telefono != "+34 600 111 222" || actualizado.Email != "javier.ms@email.com" {
		t.Fatalf("los cambios no se aplicaron: %+v", actualizado)
	}
	if !actualizado.ActualizadoEn.After(creado.ActualizadoEn) {
		t.Fatalf("esperaba que actualizado_en avanzara: antes=%v después=%v", creado.ActualizadoEn, actualizado.ActualizadoEn)
	}

	relectura, err := repo.Get(creado.ID)
	if err != nil {
		t.Fatalf("Get tras Update: %v", err)
	}
	if relectura.Telefono != "+34 600 111 222" {
		t.Fatalf("los cambios no persistieron: %+v", relectura)
	}
}

func TestInquilinosRepo_UpdateIDInexistente(t *testing.T) {
	repo := sqlite.NewInquilinosRepo(newTestDB(t))
	_, err := repo.Update(999, domain.Inquilino{NombreCompleto: "x", DocumentoIdentidad: "y"})
	if err != sqlite.ErrNotFound {
		t.Fatalf("esperaba ErrNotFound, obtuve %v", err)
	}
}

func TestInquilinosRepo_GetIDInexistente(t *testing.T) {
	repo := sqlite.NewInquilinosRepo(newTestDB(t))
	_, err := repo.Get(999)
	if err != sqlite.ErrNotFound {
		t.Fatalf("esperaba ErrNotFound, obtuve %v", err)
	}
}

func TestInquilinosRepo_List(t *testing.T) {
	repo := sqlite.NewInquilinosRepo(newTestDB(t))

	if _, err := repo.Create(domain.Inquilino{NombreCompleto: "Beatriz", DocumentoIdentidad: "1"}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, err := repo.Create(domain.Inquilino{NombreCompleto: "Andrés", DocumentoIdentidad: "2"}); err != nil {
		t.Fatalf("Create: %v", err)
	}

	lista, err := repo.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(lista) != 2 {
		t.Fatalf("esperaba 2 inquilinos, obtuve %d", len(lista))
	}
	if lista[0].NombreCompleto != "Andrés" {
		t.Fatalf("esperaba orden alfabético por nombre, obtuve %+v", lista)
	}
}

// TestInquilinosRepo_BorrarInquilinoLiberaSuHabitacion comprueba que la
// tabla habitaciones no bloquea el borrado de un inquilino ocupante
// (ON DELETE SET NULL), a diferencia de lo que hará Contrato en el Hito 3
// (ON DELETE RESTRICT).
func TestInquilinosRepo_BorrarInquilinoLiberaSuHabitacion(t *testing.T) {
	db := newTestDB(t)
	inquilinos := sqlite.NewInquilinosRepo(db)
	inmuebles := sqlite.NewInmueblesRepo(db)
	habitaciones := sqlite.NewHabitacionesRepo(db)

	inquilino, err := inquilinos.Create(domain.Inquilino{NombreCompleto: "x", DocumentoIdentidad: "y"})
	if err != nil {
		t.Fatalf("Create inquilino: %v", err)
	}
	inmueble, err := inmuebles.Create(domain.Inmueble{Nombre: "x", Direccion: "y", Tipo: domain.TipoPiso, Compartido: true})
	if err != nil {
		t.Fatalf("Create inmueble: %v", err)
	}
	habitacion, err := habitaciones.Create(domain.Habitacion{InmuebleID: inmueble.ID, Nombre: "Habitación 1"})
	if err != nil {
		t.Fatalf("Create habitación: %v", err)
	}
	if _, err := habitaciones.AsignarOcupante(habitacion.ID, &inquilino.ID); err != nil {
		t.Fatalf("AsignarOcupante: %v", err)
	}

	if _, err := db.Exec(`DELETE FROM inquilinos WHERE id = ?`, inquilino.ID); err != nil {
		t.Fatalf("no se pudo borrar el inquilino ocupante: %v", err)
	}

	liberada, err := habitaciones.Get(habitacion.ID)
	if err != nil {
		t.Fatalf("Get habitación tras borrar inquilino: %v", err)
	}
	if liberada.InquilinoID != nil {
		t.Fatalf("esperaba habitación libre tras borrar su ocupante, obtuve %+v", liberada.InquilinoID)
	}
}

// TestHabitacionesRepo_AsignarOcupanteLiberaSuHabitacionAnterior cubre la
// regla de negocio del Hito 2: una persona no ocupa dos habitaciones a la
// vez, así que asignarla a una nueva libera automáticamente la anterior.
func TestHabitacionesRepo_AsignarOcupanteLiberaSuHabitacionAnterior(t *testing.T) {
	db := newTestDB(t)
	inquilinos := sqlite.NewInquilinosRepo(db)
	inmuebles := sqlite.NewInmueblesRepo(db)
	habitaciones := sqlite.NewHabitacionesRepo(db)

	inquilino, err := inquilinos.Create(domain.Inquilino{NombreCompleto: "x", DocumentoIdentidad: "y"})
	if err != nil {
		t.Fatalf("Create inquilino: %v", err)
	}
	inmueble := crearInmuebleCompartido(t, inmuebles)
	hab1, err := habitaciones.Create(domain.Habitacion{InmuebleID: inmueble.ID, Nombre: "Habitación 1"})
	if err != nil {
		t.Fatalf("Create habitación 1: %v", err)
	}
	hab2, err := habitaciones.Create(domain.Habitacion{InmuebleID: inmueble.ID, Nombre: "Habitación 2"})
	if err != nil {
		t.Fatalf("Create habitación 2: %v", err)
	}

	if _, err := habitaciones.AsignarOcupante(hab1.ID, &inquilino.ID); err != nil {
		t.Fatalf("AsignarOcupante hab1: %v", err)
	}
	if _, err := habitaciones.AsignarOcupante(hab2.ID, &inquilino.ID); err != nil {
		t.Fatalf("AsignarOcupante hab2: %v", err)
	}

	hab1Relectura, err := habitaciones.Get(hab1.ID)
	if err != nil {
		t.Fatalf("Get hab1: %v", err)
	}
	if hab1Relectura.InquilinoID != nil {
		t.Fatalf("esperaba que hab1 quedara libre al mover el ocupante, obtuve %+v", hab1Relectura.InquilinoID)
	}
	hab2Relectura, err := habitaciones.Get(hab2.ID)
	if err != nil {
		t.Fatalf("Get hab2: %v", err)
	}
	if hab2Relectura.InquilinoID == nil || *hab2Relectura.InquilinoID != inquilino.ID {
		t.Fatalf("esperaba el inquilino en hab2, obtuve %+v", hab2Relectura.InquilinoID)
	}
}
