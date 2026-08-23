package httpapi_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/naparito/Arantxator-flat-admin/internal/domain"
)

func postInmueble(t *testing.T, srv *httptest.Server, body map[string]any) (*http.Response, map[string]any) {
	t.Helper()
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	resp, err := http.Post(srv.URL+"/api/inmuebles", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("POST /api/inmuebles: %v", err)
	}
	defer resp.Body.Close()
	var out map[string]any
	json.NewDecoder(resp.Body).Decode(&out)
	return resp, out
}

func TestAPI_CrearYLeerInmueble(t *testing.T) {
	srv, _ := newTestServer(t)

	resp, creado := postInmueble(t, srv, map[string]any{
		"nombre":    "Alcalá 145",
		"direccion": "Calle de Alcalá 145, 3ºB",
		"tipo":      "piso",
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("esperaba 201, obtuve %d: %+v", resp.StatusCode, creado)
	}
	id := creado["id"]
	if id == nil || id == float64(0) {
		t.Fatalf("esperaba un id asignado en la respuesta: %+v", creado)
	}
	if creado["estado"] != string(domain.InmuebleDisponible) {
		t.Fatalf("esperaba estado por defecto 'disponible', obtuve %v", creado["estado"])
	}

	getResp, err := http.Get(fmt.Sprintf("%s/api/inmuebles/%v", srv.URL, id))
	if err != nil {
		t.Fatalf("GET /api/inmuebles/{id}: %v", err)
	}
	defer getResp.Body.Close()
	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("esperaba 200 al leer el inmueble creado, obtuve %d", getResp.StatusCode)
	}
	var leido map[string]any
	json.NewDecoder(getResp.Body).Decode(&leido)
	if leido["nombre"] != "Alcalá 145" {
		t.Fatalf("el inmueble leído no coincide: %+v", leido)
	}
}

func TestAPI_CrearInmuebleTipoInvalido(t *testing.T) {
	srv, _ := newTestServer(t)

	resp, body := postInmueble(t, srv, map[string]any{
		"nombre":    "x",
		"direccion": "y",
		"tipo":      "castillo",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("esperaba 400 para un tipo inválido, obtuve %d: %+v", resp.StatusCode, body)
	}
	if body["error"] == nil {
		t.Fatalf("esperaba un mensaje de error en el cuerpo de la respuesta")
	}
}

func TestAPI_CrearInmuebleSinCamposObligatorios(t *testing.T) {
	srv, _ := newTestServer(t)

	resp, _ := postInmueble(t, srv, map[string]any{"tipo": "piso"})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("esperaba 400 sin nombre/dirección, obtuve %d", resp.StatusCode)
	}
}

func TestAPI_GetInmuebleInexistente(t *testing.T) {
	srv, _ := newTestServer(t)

	resp, err := http.Get(srv.URL + "/api/inmuebles/999")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("esperaba 404 para un id inexistente, obtuve %d", resp.StatusCode)
	}
}

func TestAPI_EditarInmueble(t *testing.T) {
	srv, _ := newTestServer(t)

	_, creado := postInmueble(t, srv, map[string]any{
		"nombre": "Piso en reforma", "direccion": "dir", "tipo": "piso", "estado": "en_reforma",
	})
	id := creado["id"]

	editado := map[string]any{
		"nombre": "Piso en reforma", "direccion": "dir", "tipo": "piso", "estado": "disponible", "numHabitaciones": float64(4),
	}
	b, _ := json.Marshal(editado)
	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/inmuebles/%v", srv.URL, id), bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PUT: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("esperaba 200 al editar, obtuve %d", resp.StatusCode)
	}
	var actualizado map[string]any
	json.NewDecoder(resp.Body).Decode(&actualizado)
	if actualizado["estado"] != "disponible" || actualizado["numHabitaciones"] != float64(4) {
		t.Fatalf("los cambios no se reflejan en la respuesta: %+v", actualizado)
	}
}

func TestAPI_ListarInmueblesFiltraPorEstado(t *testing.T) {
	srv, _ := newTestServer(t)

	postInmueble(t, srv, map[string]any{"nombre": "A", "direccion": "a", "tipo": "piso", "estado": "alquilado"})
	postInmueble(t, srv, map[string]any{"nombre": "B", "direccion": "b", "tipo": "piso", "estado": "disponible"})

	resp, err := http.Get(srv.URL + "/api/inmuebles?estado=alquilado")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()
	var lista []map[string]any
	json.NewDecoder(resp.Body).Decode(&lista)
	if len(lista) != 1 || lista[0]["nombre"] != "A" {
		t.Fatalf("esperaba solo el inmueble alquilado, obtuve %+v", lista)
	}
}
