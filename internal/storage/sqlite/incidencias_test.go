package sqlite_test

import (
	"testing"

	"github.com/naparito/Arantxator-flat-admin/internal/domain"
	"github.com/naparito/Arantxator-flat-admin/internal/storage/sqlite"
)

func incidenciaBase(inmuebleID int64) domain.Incidencia {
	return domain.Incidencia{
		InmuebleID: inmuebleID,
		Titulo:     "Fuga en el grifo de la cocina",
		Categoria:  "fontaneria",
		Prioridad:  domain.PrioridadAlta,
		Origen:     domain.OrigenInquilino,
	}
}

func TestIncidenciasRepo_AltaApareceEnListadoYContador(t *testing.T) {
	db := newTestDB(t)
	inmuebles := sqlite.NewInmueblesRepo(db)
	incidencias := sqlite.NewIncidenciasRepo(db)

	inmueble := crearInmuebleSimple(t, inmuebles)

	creada, err := incidencias.Create(incidenciaBase(inmueble.ID))
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if creada.Estado != domain.IncidenciaAbierta {
		t.Fatalf("una incidencia nueva debería nacer 'abierta', obtuve %q", creada.Estado)
	}
	if creada.FechaApertura.IsZero() {
		t.Fatal("la fecha de apertura debería quedar fijada al crear")
	}
	// El alta deja su propio evento con fecha.
	if len(creada.Eventos) != 1 || creada.Eventos[0].Tipo != domain.EventoAlta || creada.Eventos[0].CreadoEn.IsZero() {
		t.Fatalf("esperaba un evento de alta fechado, obtuve %+v", creada.Eventos)
	}

	lista, err := incidencias.ListByInmueble(inmueble.ID)
	if err != nil {
		t.Fatalf("ListByInmueble: %v", err)
	}
	if len(lista) != 1 || lista[0].ID != creada.ID {
		t.Fatalf("esperaba la incidencia recién creada en el listado del inmueble, obtuve %+v", lista)
	}

	n, err := incidencias.CuentaAbiertasByInmueble(inmueble.ID)
	if err != nil {
		t.Fatalf("CuentaAbiertasByInmueble: %v", err)
	}
	if n != 1 {
		t.Fatalf("esperaba 1 incidencia abierta en el contador, obtuve %d", n)
	}
}

func TestIncidenciasRepo_FlujoDeEstadoQuedaFechado(t *testing.T) {
	db := newTestDB(t)
	inmuebles := sqlite.NewInmueblesRepo(db)
	incidencias := sqlite.NewIncidenciasRepo(db)

	inmueble := crearInmuebleSimple(t, inmuebles)
	inc, err := incidencias.Create(incidenciaBase(inmueble.ID))
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	flujo := []domain.EstadoIncidencia{
		domain.IncidenciaEnProceso,
		domain.IncidenciaEsperandoProveedor,
		domain.IncidenciaResuelta,
		domain.IncidenciaCerrada,
	}
	for _, estado := range flujo {
		siguiente := inc
		siguiente.Estado = estado
		inc, err = incidencias.Update(inc.ID, siguiente)
		if err != nil {
			t.Fatalf("Update a %q: %v", estado, err)
		}
		if inc.Estado != estado {
			t.Fatalf("esperaba estado %q tras el Update, obtuve %q", estado, inc.Estado)
		}
	}

	// Un evento de alta + un cambio_estado por cada paso del flujo, todos fechados.
	cambios := 0
	for _, e := range inc.Eventos {
		if e.CreadoEn.IsZero() {
			t.Fatalf("todos los eventos deben ir fechados, este no: %+v", e)
		}
		if e.Tipo == domain.EventoCambioEstado {
			cambios++
		}
	}
	if cambios != len(flujo) {
		t.Fatalf("esperaba %d eventos de cambio de estado registrados, obtuve %d (%+v)", len(flujo), cambios, inc.Eventos)
	}
	// El último cambio deja constancia del par anterior→nuevo correcto.
	ultimo := inc.Eventos[len(inc.Eventos)-1]
	if ultimo.EstadoAnterior != domain.IncidenciaResuelta || ultimo.EstadoNuevo != domain.IncidenciaCerrada {
		t.Fatalf("el último cambio debería ser resuelta→cerrada, obtuve %s→%s", ultimo.EstadoAnterior, ultimo.EstadoNuevo)
	}

	// Al cerrar, la fecha de cierre queda sellada; ya no cuenta como abierta.
	if inc.FechaCierre == nil || inc.FechaCierre.IsZero() {
		t.Fatal("una incidencia cerrada debería tener fecha de cierre")
	}
	n, err := incidencias.CuentaAbiertasByInmueble(inmueble.ID)
	if err != nil {
		t.Fatalf("CuentaAbiertasByInmueble: %v", err)
	}
	if n != 0 {
		t.Fatalf("una incidencia cerrada no debería contar como abierta, contador = %d", n)
	}
}

func TestIncidenciasRepo_CosteACargoDeSeDistingue(t *testing.T) {
	db := newTestDB(t)
	inmuebles := sqlite.NewInmueblesRepo(db)
	incidencias := sqlite.NewIncidenciasRepo(db)
	inmueble := crearInmuebleSimple(t, inmuebles)

	aPropietario := incidenciaBase(inmueble.ID)
	aPropietario.Titulo = "Caldera"
	aPropietario.Coste = 85
	aPropietario.CosteACargoDe = domain.CostePropietario

	aInquilino := incidenciaBase(inmueble.ID)
	aInquilino.Titulo = "Cristal roto por descuido"
	aInquilino.Coste = 40
	aInquilino.CosteACargoDe = domain.CosteInquilino

	cp, err := incidencias.Create(aPropietario)
	if err != nil {
		t.Fatalf("Create propietario: %v", err)
	}
	ci, err := incidencias.Create(aInquilino)
	if err != nil {
		t.Fatalf("Create inquilino: %v", err)
	}

	leidaP, _ := incidencias.Get(cp.ID)
	leidaI, _ := incidencias.Get(ci.ID)
	if leidaP.CosteACargoDe != domain.CostePropietario || leidaP.Coste != 85 {
		t.Fatalf("esperaba coste 85 a cargo del propietario, obtuve %v / %v", leidaP.Coste, leidaP.CosteACargoDe)
	}
	if leidaI.CosteACargoDe != domain.CosteInquilino || leidaI.Coste != 40 {
		t.Fatalf("esperaba coste 40 a cargo del inquilino, obtuve %v / %v", leidaI.Coste, leidaI.CosteACargoDe)
	}
}

func TestIncidenciasRepo_SinCosteACargoDeSeGuardaNulo(t *testing.T) {
	db := newTestDB(t)
	inmuebles := sqlite.NewInmueblesRepo(db)
	incidencias := sqlite.NewIncidenciasRepo(db)
	inmueble := crearInmuebleSimple(t, inmuebles)

	// Sin coste_a_cargo_de: la columna tiene un CHECK que no admite '', así
	// que debe persistirse NULL, no una cadena vacía (si no, el INSERT falla).
	sinCargo := incidenciaBase(inmueble.ID)
	sinCargo.CosteACargoDe = ""
	sinCargo.Origen = ""
	creada, err := incidencias.Create(sinCargo)
	if err != nil {
		t.Fatalf("Create sin coste a cargo de: %v", err)
	}
	if creada.CosteACargoDe != "" {
		t.Fatalf("esperaba coste a cargo de vacío, obtuve %q", creada.CosteACargoDe)
	}
}

func TestIncidenciasRepo_BorrarInmuebleArrastraIncidenciasYEventos(t *testing.T) {
	db := newTestDB(t)
	inmuebles := sqlite.NewInmueblesRepo(db)
	incidencias := sqlite.NewIncidenciasRepo(db)
	inmueble := crearInmuebleSimple(t, inmuebles)

	inc, err := incidencias.Create(incidenciaBase(inmueble.ID))
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if _, err := db.Exec(`DELETE FROM inmuebles WHERE id = ?`, inmueble.ID); err != nil {
		t.Fatalf("DELETE inmueble: %v", err)
	}
	if _, err := incidencias.Get(inc.ID); err != sqlite.ErrNotFound {
		t.Fatalf("esperaba que la incidencia se borrara en cascada con su inmueble, err = %v", err)
	}
	var eventos int
	db.QueryRow(`SELECT COUNT(*) FROM incidencia_eventos WHERE incidencia_id = ?`, inc.ID).Scan(&eventos)
	if eventos != 0 {
		t.Fatalf("los eventos de la incidencia deberían borrarse en cascada, quedan %d", eventos)
	}
}
