package httpapi_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func postInquilino(t *testing.T, srv *httptest.Server, body map[string]any) (*http.Response, map[string]any) {
	t.Helper()
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	resp, err := http.Post(srv.URL+"/api/inquilinos", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("POST /api/inquilinos: %v", err)
	}
	defer resp.Body.Close()
	var out map[string]any
	json.NewDecoder(resp.Body).Decode(&out)
	return resp, out
}

func TestAPI_CrearYLeerInquilino(t *testing.T) {
	srv, _ := newTestServer(t)

	resp, creado := postInquilino(t, srv, map[string]any{
		"nombreCompleto":     "Laura Fernández Ruiz",
		"documentoIdentidad": "45123456M",
		"iban":               "ES9121000000000000001234",
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("esperaba 201, obtuve %d: %+v", resp.StatusCode, creado)
	}
	id := creado["id"]
	if id == nil || id == float64(0) {
		t.Fatalf("esperaba un id asignado en la respuesta: %+v", creado)
	}
	if creado["iban"] != "ES9121000000000000001234" {
		t.Fatalf("esperaba el IBAN completo en la respuesta de la API (el enmascarado es solo de presentación), obtuve %v", creado["iban"])
	}

	getResp, err := http.Get(fmt.Sprintf("%s/api/inquilinos/%v", srv.URL, id))
	if err != nil {
		t.Fatalf("GET /api/inquilinos/{id}: %v", err)
	}
	defer getResp.Body.Close()
	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("esperaba 200 al leer el inquilino creado, obtuve %d", getResp.StatusCode)
	}
	var leido map[string]any
	json.NewDecoder(getResp.Body).Decode(&leido)
	if leido["nombreCompleto"] != "Laura Fernández Ruiz" {
		t.Fatalf("el inquilino leído no coincide: %+v", leido)
	}
}

func TestAPI_CrearInquilinoSinCamposObligatorios(t *testing.T) {
	srv, _ := newTestServer(t)

	resp, _ := postInquilino(t, srv, map[string]any{"telefono": "600111222"})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("esperaba 400 sin nombreCompleto/documentoIdentidad, obtuve %d", resp.StatusCode)
	}
}

func TestAPI_GetInquilinoInexistente(t *testing.T) {
	srv, _ := newTestServer(t)

	resp, err := http.Get(srv.URL + "/api/inquilinos/999")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("esperaba 404 para un id inexistente, obtuve %d", resp.StatusCode)
	}
}

func TestAPI_EditarInquilino(t *testing.T) {
	srv, _ := newTestServer(t)

	_, creado := postInquilino(t, srv, map[string]any{"nombreCompleto": "Javier Martín Soto", "documentoIdentidad": "Y1234567L"})
	id := creado["id"]

	editado := map[string]any{
		"nombreCompleto": "Javier Martín Soto", "documentoIdentidad": "Y1234567L", "telefono": "+34 600 111 222",
	}
	b, _ := json.Marshal(editado)
	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/inquilinos/%v", srv.URL, id), bytes.NewReader(b))
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
	if actualizado["telefono"] != "+34 600 111 222" {
		t.Fatalf("los cambios no se reflejan en la respuesta: %+v", actualizado)
	}
}

func TestAPI_ListarInquilinos(t *testing.T) {
	srv, _ := newTestServer(t)

	postInquilino(t, srv, map[string]any{"nombreCompleto": "Beatriz", "documentoIdentidad": "1"})
	postInquilino(t, srv, map[string]any{"nombreCompleto": "Andrés", "documentoIdentidad": "2"})

	resp, err := http.Get(srv.URL + "/api/inquilinos")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()
	var lista []map[string]any
	json.NewDecoder(resp.Body).Decode(&lista)
	if len(lista) != 2 {
		t.Fatalf("esperaba 2 inquilinos, obtuve %+v", lista)
	}
}

func TestAPI_AsignarYQuitarOcupanteHabitacion(t *testing.T) {
	srv, _ := newTestServer(t)

	_, inmueble := postInmueble(t, srv, map[string]any{"nombre": "x", "direccion": "y", "tipo": "piso", "compartido": true})
	_, hab := postHabitacion(t, srv.URL, inmueble["id"], map[string]any{"nombre": "Habitación 1"})
	_, inquilino := postInquilino(t, srv, map[string]any{"nombreCompleto": "x", "documentoIdentidad": "y"})

	asignar := func(inquilinoID any) (*http.Response, map[string]any) {
		b, _ := json.Marshal(map[string]any{"inquilinoId": inquilinoID})
		req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/habitaciones/%v/ocupante", srv.URL, hab["id"]), bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("PUT ocupante: %v", err)
		}
		defer resp.Body.Close()
		var out map[string]any
		json.NewDecoder(resp.Body).Decode(&out)
		return resp, out
	}

	resp, actualizada := asignar(inquilino["id"])
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("esperaba 200 al asignar ocupante, obtuve %d: %+v", resp.StatusCode, actualizada)
	}
	if actualizada["inquilinoId"] != inquilino["id"] {
		t.Fatalf("esperaba el inquilino asignado, obtuve %+v", actualizada)
	}

	resp2, liberada := asignar(nil)
	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("esperaba 200 al quitar ocupante, obtuve %d: %+v", resp2.StatusCode, liberada)
	}
	if liberada["inquilinoId"] != nil {
		t.Fatalf("esperaba habitación libre, obtuve %+v", liberada)
	}
}

func TestAPI_AsignarOcupanteInquilinoInexistente(t *testing.T) {
	srv, _ := newTestServer(t)

	_, inmueble := postInmueble(t, srv, map[string]any{"nombre": "x", "direccion": "y", "tipo": "piso", "compartido": true})
	_, hab := postHabitacion(t, srv.URL, inmueble["id"], map[string]any{"nombre": "Habitación 1"})

	b, _ := json.Marshal(map[string]any{"inquilinoId": 999})
	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/habitaciones/%v/ocupante", srv.URL, hab["id"]), bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PUT: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("esperaba 404 para un inquilino inexistente, obtuve %d", resp.StatusCode)
	}
}

func TestAPI_AsignarOcupanteHabitacionInexistente(t *testing.T) {
	srv, _ := newTestServer(t)

	b, _ := json.Marshal(map[string]any{"inquilinoId": nil})
	req, _ := http.NewRequest(http.MethodPut, srv.URL+"/api/habitaciones/999/ocupante", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PUT: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("esperaba 404 para una habitación inexistente, obtuve %d", resp.StatusCode)
	}
}

func TestAPI_AsignarOcupanteYaOcupanteDeOtraHabitacionLaLibera(t *testing.T) {
	srv, _ := newTestServer(t)

	_, inmueble := postInmueble(t, srv, map[string]any{"nombre": "x", "direccion": "y", "tipo": "piso", "compartido": true})
	_, hab1 := postHabitacion(t, srv.URL, inmueble["id"], map[string]any{"nombre": "Habitación 1"})
	_, hab2 := postHabitacion(t, srv.URL, inmueble["id"], map[string]any{"nombre": "Habitación 2"})
	_, inquilino := postInquilino(t, srv, map[string]any{"nombreCompleto": "x", "documentoIdentidad": "y"})

	put := func(habID, inquilinoID any) *http.Response {
		b, _ := json.Marshal(map[string]any{"inquilinoId": inquilinoID})
		req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/habitaciones/%v/ocupante", srv.URL, habID), bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("PUT: %v", err)
		}
		return resp
	}

	put(hab1["id"], inquilino["id"]).Body.Close()
	put(hab2["id"], inquilino["id"]).Body.Close()

	getResp, err := http.Get(fmt.Sprintf("%s/api/habitaciones/%v", srv.URL, hab1["id"]))
	if err != nil {
		t.Fatalf("GET hab1: %v", err)
	}
	defer getResp.Body.Close()
	var hab1Actual map[string]any
	json.NewDecoder(getResp.Body).Decode(&hab1Actual)
	if hab1Actual["inquilinoId"] != nil {
		t.Fatalf("esperaba que hab1 quedara libre tras mover el ocupante a hab2, obtuve %+v", hab1Actual)
	}
}
