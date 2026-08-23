package sqlite_test

import (
	"testing"
	"time"

	"github.com/naparito/Arantxator-flat-admin/internal/domain"
	"github.com/naparito/Arantxator-flat-admin/internal/storage/sqlite"
)

func TestInmueblesRepo_CreateConTodosLosCampos(t *testing.T) {
	repo := sqlite.NewInmueblesRepo(newTestDB(t))
	caducidad := time.Date(2029, 3, 15, 0, 0, 0, 0, time.UTC)

	entrada := domain.Inmueble{
		Nombre:                         "Bravo Murillo 210",
		Direccion:                      "Calle Bravo Murillo 210, Bajo A",
		ReferenciaCatastral:            "1234567AB1234C0001XY",
		CodigoPostal:                   "28020",
		Ciudad:                         "Madrid",
		Provincia:                      "Madrid",
		Tipo:                           domain.TipoPiso,
		M2Construidos:                  95,
		M2Utiles:                       88,
		NumHabitaciones:                3,
		NumBanos:                       2,
		Planta:                         "Bajo",
		Ascensor:                       true,
		Amueblado:                      true,
		AnioConstruccion:               1975,
		CertificadoEnergeticoLetra:     "D",
		CertificadoEnergeticoCaducidad: &caducidad,
		Estado:                         domain.InmuebleAlquilado,
		Suministros: domain.Suministros{
			Luz:      domain.Suministro{Compania: "Iberdrola", NumeroContrato: "LUZ-1", Titular: "Marta Gómez"},
			Agua:     domain.Suministro{Compania: "Canal de Isabel II", NumeroContrato: "AGU-1", Titular: "Marta Gómez"},
			Gas:      domain.Suministro{Compania: "Naturgy", NumeroContrato: "GAS-1", Titular: "Marta Gómez"},
			Internet: domain.Suministro{Compania: "Movistar", NumeroContrato: "NET-1", Titular: "Marta Gómez"},
		},
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

	if leido.Nombre != entrada.Nombre ||
		leido.Direccion != entrada.Direccion ||
		leido.ReferenciaCatastral != entrada.ReferenciaCatastral ||
		leido.CodigoPostal != entrada.CodigoPostal ||
		leido.Ciudad != entrada.Ciudad ||
		leido.Provincia != entrada.Provincia ||
		leido.Tipo != entrada.Tipo ||
		leido.M2Construidos != entrada.M2Construidos ||
		leido.M2Utiles != entrada.M2Utiles ||
		leido.NumHabitaciones != entrada.NumHabitaciones ||
		leido.NumBanos != entrada.NumBanos ||
		leido.Planta != entrada.Planta ||
		leido.Ascensor != entrada.Ascensor ||
		leido.Amueblado != entrada.Amueblado ||
		leido.AnioConstruccion != entrada.AnioConstruccion ||
		leido.CertificadoEnergeticoLetra != entrada.CertificadoEnergeticoLetra ||
		leido.Estado != entrada.Estado ||
		leido.Suministros != entrada.Suministros {
		t.Fatalf("el inmueble leído no coincide con el creado:\n  creado=%+v\n  leído =%+v", entrada, leido)
	}
	if leido.CertificadoEnergeticoCaducidad == nil || !leido.CertificadoEnergeticoCaducidad.Equal(caducidad) {
		t.Fatalf("fecha de caducidad del certificado no coincide: %v", leido.CertificadoEnergeticoCaducidad)
	}
	if leido.CreadoEn.IsZero() || leido.ActualizadoEn.IsZero() {
		t.Fatalf("esperaba creado_en/actualizado_en no vacíos")
	}
}

func TestInmueblesRepo_CreateSoloConCamposObligatorios(t *testing.T) {
	repo := sqlite.NewInmueblesRepo(newTestDB(t))

	creado, err := repo.Create(domain.Inmueble{
		Nombre:    "Alcalá 145",
		Direccion: "Calle de Alcalá 145, 3ºB",
		Tipo:      domain.TipoPiso,
	})
	if err != nil {
		t.Fatalf("Create con campos mínimos: %v", err)
	}

	leido, err := repo.Get(creado.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if leido.Estado != domain.InmuebleDisponible {
		t.Fatalf("esperaba estado por defecto 'disponible', obtuve %q", leido.Estado)
	}
	if leido.ReferenciaCatastral != "" || leido.CodigoPostal != "" || leido.CertificadoEnergeticoLetra != "" {
		t.Fatalf("esperaba campos opcionales vacíos, obtuve %+v", leido)
	}
	if leido.CertificadoEnergeticoCaducidad != nil {
		t.Fatalf("esperaba fecha de caducidad nula, obtuve %v", leido.CertificadoEnergeticoCaducidad)
	}
}

func TestInmueblesRepo_Update(t *testing.T) {
	repo := sqlite.NewInmueblesRepo(newTestDB(t))

	creado, err := repo.Create(domain.Inmueble{
		Nombre:    "Piso en reforma",
		Direccion: "Paseo de la Castellana 89, 7ºA",
		Tipo:      domain.TipoPiso,
		Estado:    domain.InmuebleEnReforma,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	time.Sleep(1100 * time.Millisecond) // CURRENT_TIMESTAMP de SQLite tiene resolución de segundo

	editado := creado
	editado.Estado = domain.InmuebleDisponible
	editado.NumHabitaciones = 4

	actualizado, err := repo.Update(creado.ID, editado)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if actualizado.Estado != domain.InmuebleDisponible || actualizado.NumHabitaciones != 4 {
		t.Fatalf("los cambios no se aplicaron: %+v", actualizado)
	}
	if !actualizado.ActualizadoEn.After(creado.ActualizadoEn) {
		t.Fatalf("esperaba que actualizado_en avanzara: antes=%v después=%v", creado.ActualizadoEn, actualizado.ActualizadoEn)
	}

	relectura, err := repo.Get(creado.ID)
	if err != nil {
		t.Fatalf("Get tras Update: %v", err)
	}
	if relectura.Estado != domain.InmuebleDisponible || relectura.NumHabitaciones != 4 {
		t.Fatalf("los cambios no persistieron: %+v", relectura)
	}
}

func TestInmueblesRepo_UpdateIDInexistente(t *testing.T) {
	repo := sqlite.NewInmueblesRepo(newTestDB(t))
	_, err := repo.Update(999, domain.Inmueble{Nombre: "x", Direccion: "y", Tipo: domain.TipoPiso})
	if err != sqlite.ErrNotFound {
		t.Fatalf("esperaba ErrNotFound, obtuve %v", err)
	}
}

func TestInmueblesRepo_GetIDInexistente(t *testing.T) {
	repo := sqlite.NewInmueblesRepo(newTestDB(t))
	_, err := repo.Get(999)
	if err != sqlite.ErrNotFound {
		t.Fatalf("esperaba ErrNotFound, obtuve %v", err)
	}
}

func TestInmueblesRepo_ListFiltraPorEstado(t *testing.T) {
	repo := sqlite.NewInmueblesRepo(newTestDB(t))

	crear := func(nombre string, estado domain.EstadoInmueble) {
		if _, err := repo.Create(domain.Inmueble{Nombre: nombre, Direccion: "dir " + nombre, Tipo: domain.TipoPiso, Estado: estado}); err != nil {
			t.Fatalf("Create %s: %v", nombre, err)
		}
	}
	crear("A", domain.InmuebleAlquilado)
	crear("B", domain.InmuebleDisponible)
	crear("C", domain.InmuebleAlquilado)
	crear("D", domain.InmuebleEnReforma)

	alquilados, err := repo.List(domain.InmuebleAlquilado)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(alquilados) != 2 {
		t.Fatalf("esperaba 2 inmuebles alquilados, obtuve %d: %+v", len(alquilados), alquilados)
	}
	for _, m := range alquilados {
		if m.Estado != domain.InmuebleAlquilado {
			t.Fatalf("List devolvió un inmueble con estado %q al filtrar por 'alquilado'", m.Estado)
		}
	}

	todos, err := repo.List("")
	if err != nil {
		t.Fatalf("List sin filtro: %v", err)
	}
	if len(todos) != 4 {
		t.Fatalf("esperaba 4 inmuebles sin filtro, obtuve %d", len(todos))
	}
}

func TestInmueblesRepo_Archivar(t *testing.T) {
	repo := sqlite.NewInmueblesRepo(newTestDB(t))
	creado, err := repo.Create(domain.Inmueble{Nombre: "x", Direccion: "y", Tipo: domain.TipoCasa})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := repo.Archivar(creado.ID); err != nil {
		t.Fatalf("Archivar: %v", err)
	}
	leido, err := repo.Get(creado.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if leido.Estado != domain.InmuebleFueraDeServicio {
		t.Fatalf("esperaba estado 'fuera_de_servicio', obtuve %q", leido.Estado)
	}
}
