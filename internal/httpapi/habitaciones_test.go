package httpapi_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func postHabitacion(t *testing.T, baseURL string, inmuebleID any, body map[string]any) (*http.Response, map[string]any) {
	t.Helper()
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	resp, err := http.Post(fmt.Sprintf("%s/api/inmuebles/%v/habitaciones", baseURL, inmuebleID), "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("POST habitaciones: %v", err)
	}
	defer resp.Body.Close()
	var out map[string]any
	json.NewDecoder(resp.Body).Decode(&out)
	return resp, out
}

func TestAPI_CrearInmuebleCompartidoYHabitaciones(t *testing.T) {
	srv, _ := newTestServer(t)

	_, inmueble := postInmueble(t, srv, map[string]any{
		"nombre": "Bravo Murillo 210", "direccion": "dir", "tipo": "piso", "compartido": true,
	})
	if inmueble["compartido"] != true {
		t.Fatalf("esperaba compartido=true en la respuesta de alta, obtuve %+v", inmueble)
	}
	inmuebleID := inmueble["id"]

	resp, hab := postHabitacion(t, srv.URL, inmuebleID, map[string]any{"nombre": "Habitación 1", "m2": float64(12), "tieneBano": true})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("esperaba 201 al crear la habitación, obtuve %d: %+v", resp.StatusCode, hab)
	}
	if hab["nombre"] != "Habitación 1" || hab["tieneBano"] != true {
		t.Fatalf("habitación creada inesperada: %+v", hab)
	}
	if hab["inquilinoId"] != nil {
		t.Fatalf("esperaba inquilinoId nulo al crear, obtuve %v", hab["inquilinoId"])
	}

	listaResp, err := http.Get(fmt.Sprintf("%s/api/inmuebles/%v/habitaciones", srv.URL, inmuebleID))
	if err != nil {
		t.Fatalf("GET habitaciones: %v", err)
	}
	defer listaResp.Body.Close()
	var lista []map[string]any
	json.NewDecoder(listaResp.Body).Decode(&lista)
	if len(lista) != 1 || lista[0]["nombre"] != "Habitación 1" {
		t.Fatalf("esperaba la habitación recién creada en el listado, obtuve %+v", lista)
	}
}

func TestAPI_HabitacionesDeInmuebleInexistente(t *testing.T) {
	srv, _ := newTestServer(t)

	resp, err := http.Get(srv.URL + "/api/inmuebles/999/habitaciones")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("esperaba 404, obtuve %d", resp.StatusCode)
	}

	resp2, _ := postHabitacion(t, srv.URL, 999, map[string]any{"nombre": "x"})
	if resp2.StatusCode != http.StatusNotFound {
		t.Fatalf("esperaba 404 al crear en un inmueble inexistente, obtuve %d", resp2.StatusCode)
	}
}

func TestAPI_CrearHabitacionSinNombre(t *testing.T) {
	srv, _ := newTestServer(t)
	_, inmueble := postInmueble(t, srv, map[string]any{"nombre": "x", "direccion": "y", "tipo": "piso", "compartido": true})

	resp, _ := postHabitacion(t, srv.URL, inmueble["id"], map[string]any{})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("esperaba 400 sin nombre, obtuve %d", resp.StatusCode)
	}
}

func TestAPI_EditarYBorrarHabitacion(t *testing.T) {
	srv, _ := newTestServer(t)
	_, inmueble := postInmueble(t, srv, map[string]any{"nombre": "x", "direccion": "y", "tipo": "piso", "compartido": true})
	_, hab := postHabitacion(t, srv.URL, inmueble["id"], map[string]any{"nombre": "Habitación 1"})
	habID := hab["id"]

	editado := map[string]any{"nombre": "Doble exterior", "m2": float64(16), "amueblada": true}
	b, _ := json.Marshal(editado)
	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/habitaciones/%v", srv.URL, habID), bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PUT: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("esperaba 200 al editar, obtuve %d", resp.StatusCode)
	}
	var actualizada map[string]any
	json.NewDecoder(resp.Body).Decode(&actualizada)
	if actualizada["nombre"] != "Doble exterior" || actualizada["amueblada"] != true {
		t.Fatalf("los cambios no se reflejan: %+v", actualizada)
	}

	delReq, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/habitaciones/%v", srv.URL, habID), nil)
	delResp, err := http.DefaultClient.Do(delReq)
	if err != nil {
		t.Fatalf("DELETE: %v", err)
	}
	defer delResp.Body.Close()
	if delResp.StatusCode != http.StatusNoContent {
		t.Fatalf("esperaba 204 al borrar, obtuve %d", delResp.StatusCode)
	}

	getResp, err := http.Get(fmt.Sprintf("%s/api/habitaciones/%v", srv.URL, habID))
	if err != nil {
		t.Fatalf("GET tras borrar: %v", err)
	}
	defer getResp.Body.Close()
	if getResp.StatusCode != http.StatusNotFound {
		t.Fatalf("esperaba 404 tras borrar, obtuve %d", getResp.StatusCode)
	}
}
