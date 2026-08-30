package httpapi_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func postContrato(t *testing.T, srv *httptest.Server, body map[string]any) (*http.Response, map[string]any) {
	t.Helper()
	b, _ := json.Marshal(body)
	resp, err := http.Post(srv.URL+"/api/contratos", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("POST /api/contratos: %v", err)
	}
	defer resp.Body.Close()
	var out map[string]any
	json.NewDecoder(resp.Body).Decode(&out)
	return resp, out
}

// contratoBody arma el cuerpo mínimo de un contrato para un inmueble y un
// inquilino ya creados.
func contratoBody(inmuebleID, inquilinoID any) map[string]any {
	return map[string]any{
		"inmuebleId":   inmuebleID,
		"inquilinoIds": []any{inquilinoID},
		"fechaFirma":   "2026-02-01",
		"fechaInicio":  "2026-02-01",
		"fechaFin":     "2031-01-31",
		"rentaMensual": 980,
		"diaPago":      5,
	}
}

func crearInmuebleYInquilino(t *testing.T, srv *httptest.Server, compartido bool) (map[string]any, map[string]any) {
	t.Helper()
	_, inmueble := postInmueble(t, srv, map[string]any{
		"nombre": "Alcalá 145", "direccion": "Calle de Alcalá 145, 3ºB", "tipo": "piso", "compartido": compartido,
	})
	_, inquilino := postInquilino(t, srv, map[string]any{"nombreCompleto": "Laura Fernández Ruiz", "documentoIdentidad": "45123456M"})
	return inmueble, inquilino
}

func TestAPI_CrearYLeerContrato(t *testing.T) {
	srv, _ := newTestServer(t)
	inmueble, inquilino := crearInmuebleYInquilino(t, srv, false)

	resp, creado := postContrato(t, srv, contratoBody(inmueble["id"], inquilino["id"]))
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("esperaba 201, obtuve %d: %+v", resp.StatusCode, creado)
	}
	// Las fechas viajan en formato AAAA-MM-DD (lo que manda un <input type="date">).
	if creado["fechaFirma"] != "2026-02-01" {
		t.Fatalf("esperaba la fecha de firma tal cual, obtuve %v", creado["fechaFirma"])
	}
	// La fecha límite de depósito de fianza es exactamente firma + 30 días.
	if creado["fechaLimiteDepositoFianza"] != "2026-03-03" {
		t.Fatalf("esperaba fechaLimiteDepositoFianza = 2026-03-03 (firma + 30 días), obtuve %v", creado["fechaLimiteDepositoFianza"])
	}
	if creado["estado"] != "activo" {
		t.Fatalf("un contrato con fin lejano debería estar 'activo', obtuve %v", creado["estado"])
	}

	getResp, err := http.Get(fmt.Sprintf("%s/api/contratos/%v", srv.URL, creado["id"]))
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer getResp.Body.Close()
	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("esperaba 200 al leer el contrato, obtuve %d", getResp.StatusCode)
	}
	var leido map[string]any
	json.NewDecoder(getResp.Body).Decode(&leido)
	ids, _ := leido["inquilinoIds"].([]any)
	if len(ids) != 1 {
		t.Fatalf("esperaba 1 co-arrendatario, obtuve %+v", leido["inquilinoIds"])
	}
}

func TestAPI_CrearContratoConTresInquilinos(t *testing.T) {
	srv, _ := newTestServer(t)
	_, inmueble := postInmueble(t, srv, map[string]any{"nombre": "x", "direccion": "y", "tipo": "piso"})
	_, a := postInquilino(t, srv, map[string]any{"nombreCompleto": "Ana", "documentoIdentidad": "1"})
	_, b := postInquilino(t, srv, map[string]any{"nombreCompleto": "Bruno", "documentoIdentidad": "2"})
	_, c := postInquilino(t, srv, map[string]any{"nombreCompleto": "Carla", "documentoIdentidad": "3"})

	body := contratoBody(inmueble["id"], a["id"])
	body["inquilinoIds"] = []any{a["id"], b["id"], c["id"]}
	resp, creado := postContrato(t, srv, body)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("esperaba 201, obtuve %d: %+v", resp.StatusCode, creado)
	}
	ids, _ := creado["inquilinoIds"].([]any)
	if len(ids) != 3 {
		t.Fatalf("esperaba 3 co-arrendatarios, obtuve %+v", creado["inquilinoIds"])
	}

	// El histórico de cada inquilino recoge el contrato.
	for _, inq := range []map[string]any{a, b, c} {
		r, err := http.Get(fmt.Sprintf("%s/api/inquilinos/%v/contratos", srv.URL, inq["id"]))
		if err != nil {
			t.Fatalf("GET histórico: %v", err)
		}
		var hist []map[string]any
		json.NewDecoder(r.Body).Decode(&hist)
		r.Body.Close()
		if len(hist) != 1 {
			t.Fatalf("inquilino %v: esperaba 1 contrato en su histórico, obtuve %+v", inq["id"], hist)
		}
	}
}

func TestAPI_CrearContratoInmuebleCompartidoSinHabitacion(t *testing.T) {
	srv, _ := newTestServer(t)
	inmueble, inquilino := crearInmuebleYInquilino(t, srv, true)

	resp, body := postContrato(t, srv, contratoBody(inmueble["id"], inquilino["id"]))
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("esperaba 400 al crear un contrato sin habitación en un inmueble compartido, obtuve %d: %+v", resp.StatusCode, body)
	}
}

func TestAPI_ContratosCompartidoDosHabitacionesConvivenSinSolapar(t *testing.T) {
	srv, _ := newTestServer(t)
	inmueble, a := crearInmuebleYInquilino(t, srv, true)
	_, b := postInquilino(t, srv, map[string]any{"nombreCompleto": "Bruno", "documentoIdentidad": "2"})
	_, hab1 := postHabitacion(t, srv.URL, inmueble["id"], map[string]any{"nombre": "Habitación 1"})
	_, hab2 := postHabitacion(t, srv.URL, inmueble["id"], map[string]any{"nombre": "Habitación 2"})

	c1 := contratoBody(inmueble["id"], a["id"])
	c1["habitacionId"] = hab1["id"]
	if resp, out := postContrato(t, srv, c1); resp.StatusCode != http.StatusCreated {
		t.Fatalf("esperaba 201 para el contrato de la habitación 1, obtuve %d: %+v", resp.StatusCode, out)
	}

	c2 := contratoBody(inmueble["id"], b["id"])
	c2["habitacionId"] = hab2["id"]
	if resp, out := postContrato(t, srv, c2); resp.StatusCode != http.StatusCreated {
		t.Fatalf("esperaba 201 para el contrato de la habitación 2 (libre), obtuve %d: %+v", resp.StatusCode, out)
	}

	// Segundo contrato activo sobre la habitación 1 -> solapamiento, se rechaza.
	c3 := contratoBody(inmueble["id"], b["id"])
	c3["habitacionId"] = hab1["id"]
	if resp, out := postContrato(t, srv, c3); resp.StatusCode != http.StatusConflict {
		t.Fatalf("esperaba 409 por solapamiento en la habitación 1, obtuve %d: %+v", resp.StatusCode, out)
	}

	// Ambos contratos conviven activos.
	r, _ := http.Get(srv.URL + "/api/contratos")
	var lista []map[string]any
	json.NewDecoder(r.Body).Decode(&lista)
	r.Body.Close()
	activos := 0
	for _, c := range lista {
		if c["estado"] == "activo" {
			activos++
		}
	}
	if activos != 2 {
		t.Fatalf("esperaba 2 contratos activos a la vez en el inmueble compartido, obtuve %d: %+v", activos, lista)
	}
}

func TestAPI_ContratoNoCompartidoSegundoActivoSeRechaza(t *testing.T) {
	srv, _ := newTestServer(t)
	inmueble, a := crearInmuebleYInquilino(t, srv, false)
	_, b := postInquilino(t, srv, map[string]any{"nombreCompleto": "Bruno", "documentoIdentidad": "2"})

	if resp, out := postContrato(t, srv, contratoBody(inmueble["id"], a["id"])); resp.StatusCode != http.StatusCreated {
		t.Fatalf("esperaba 201 para el primer contrato, obtuve %d: %+v", resp.StatusCode, out)
	}
	if resp, out := postContrato(t, srv, contratoBody(inmueble["id"], b["id"])); resp.StatusCode != http.StatusConflict {
		t.Fatalf("esperaba 409 para un segundo contrato activo sobre el mismo inmueble, obtuve %d: %+v", resp.StatusCode, out)
	}
}

func TestAPI_ContratoProximoAVencer(t *testing.T) {
	srv, _ := newTestServer(t)
	inmueble, inquilino := crearInmuebleYInquilino(t, srv, false)

	body := contratoBody(inmueble["id"], inquilino["id"])
	body["fechaInicio"] = "2022-01-01"
	body["fechaFin"] = time.Now().AddDate(0, 0, 20).Format("2006-01-02") // dentro de la ventana de 60 días

	resp, creado := postContrato(t, srv, body)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("esperaba 201, obtuve %d: %+v", resp.StatusCode, creado)
	}
	if creado["estado"] != "proximo_a_vencer" {
		t.Fatalf("un contrato que vence en 20 días debería estar 'proximo_a_vencer', obtuve %v", creado["estado"])
	}
}

func TestAPI_ContratoVencido(t *testing.T) {
	srv, _ := newTestServer(t)
	inmueble, inquilino := crearInmuebleYInquilino(t, srv, false)

	body := contratoBody(inmueble["id"], inquilino["id"])
	body["fechaInicio"] = "2019-01-01"
	body["fechaFin"] = "2020-01-01"

	resp, creado := postContrato(t, srv, body)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("esperaba 201, obtuve %d: %+v", resp.StatusCode, creado)
	}
	if creado["estado"] != "vencido" {
		t.Fatalf("un contrato con fin en 2020 debería estar 'vencido', obtuve %v", creado["estado"])
	}
}

func TestAPI_OcupacionInmuebleCompartidoEnListadoYFicha(t *testing.T) {
	srv, _ := newTestServer(t)
	_, compartido := postInmueble(t, srv, map[string]any{"nombre": "Compartido", "direccion": "dir", "tipo": "piso", "compartido": true})
	_, noCompartido := postInmueble(t, srv, map[string]any{"nombre": "Entero", "direccion": "dir2", "tipo": "piso"})
	_, hab1 := postHabitacion(t, srv.URL, compartido["id"], map[string]any{"nombre": "Habitación 1"})
	_, hab2 := postHabitacion(t, srv.URL, compartido["id"], map[string]any{"nombre": "Habitación 2"})
	postHabitacion(t, srv.URL, compartido["id"], map[string]any{"nombre": "Habitación 3"})

	_, a := postInquilino(t, srv, map[string]any{"nombreCompleto": "Ana", "documentoIdentidad": "1"})
	_, b := postInquilino(t, srv, map[string]any{"nombreCompleto": "Bruno", "documentoIdentidad": "2"})
	for _, par := range []struct{ hab, inq map[string]any }{{hab1, a}, {hab2, b}} {
		body := contratoBody(compartido["id"], par.inq["id"])
		body["habitacionId"] = par.hab["id"]
		if resp, out := postContrato(t, srv, body); resp.StatusCode != http.StatusCreated {
			t.Fatalf("Create contrato: %d %+v", resp.StatusCode, out)
		}
	}

	// Listado
	r, _ := http.Get(srv.URL + "/api/inmuebles")
	var lista []map[string]any
	json.NewDecoder(r.Body).Decode(&lista)
	r.Body.Close()
	for _, m := range lista {
		if fmt.Sprint(m["id"]) == fmt.Sprint(compartido["id"]) {
			oc, ok := m["ocupacion"].(map[string]any)
			if !ok {
				t.Fatalf("esperaba un objeto 'ocupacion' en el inmueble compartido, obtuve %+v", m["ocupacion"])
			}
			if oc["porcentaje"] != float64(67) || oc["habitacionesOcupadas"] != float64(2) || oc["habitacionesTotales"] != float64(3) {
				t.Fatalf("esperaba 2/3 = 67%%, obtuve %+v", oc)
			}
		}
		if fmt.Sprint(m["id"]) == fmt.Sprint(noCompartido["id"]) {
			if _, tiene := m["ocupacion"]; tiene {
				t.Fatalf("un inmueble no compartido no debería llevar 'ocupacion', obtuve %+v", m["ocupacion"])
			}
		}
	}

	// Ficha
	r2, _ := http.Get(fmt.Sprintf("%s/api/inmuebles/%v", srv.URL, compartido["id"]))
	var ficha map[string]any
	json.NewDecoder(r2.Body).Decode(&ficha)
	r2.Body.Close()
	oc, ok := ficha["ocupacion"].(map[string]any)
	if !ok || oc["porcentaje"] != float64(67) {
		t.Fatalf("la ficha del inmueble compartido debería mostrar el 67%% de ocupación, obtuve %+v", ficha["ocupacion"])
	}
}

func TestAPI_ContratoCompartidoNoTocaEstadoInmueble(t *testing.T) {
	srv, _ := newTestServer(t)
	inmueble, inquilino := crearInmuebleYInquilino(t, srv, true)
	_, hab := postHabitacion(t, srv.URL, inmueble["id"], map[string]any{"nombre": "Habitación 1"})

	body := contratoBody(inmueble["id"], inquilino["id"])
	body["habitacionId"] = hab["id"]
	if resp, out := postContrato(t, srv, body); resp.StatusCode != http.StatusCreated {
		t.Fatalf("Create: %d %+v", resp.StatusCode, out)
	}

	r, _ := http.Get(fmt.Sprintf("%s/api/inmuebles/%v", srv.URL, inmueble["id"]))
	var m map[string]any
	json.NewDecoder(r.Body).Decode(&m)
	r.Body.Close()
	if m["estado"] != "disponible" {
		t.Fatalf("el estado de un inmueble compartido no debe cambiar al activar un contrato, obtuve %v", m["estado"])
	}
}

func TestAPI_ContratoNoCompartidoAlquilaElInmueble(t *testing.T) {
	srv, _ := newTestServer(t)
	inmueble, inquilino := crearInmuebleYInquilino(t, srv, false)

	body := contratoBody(inmueble["id"], inquilino["id"])
	body["fechaFin"] = time.Now().AddDate(2, 0, 0).Format("2006-01-02")
	resp, creado := postContrato(t, srv, body)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Create: %d %+v", resp.StatusCode, creado)
	}

	get := func() string {
		r, _ := http.Get(fmt.Sprintf("%s/api/inmuebles/%v", srv.URL, inmueble["id"]))
		var m map[string]any
		json.NewDecoder(r.Body).Decode(&m)
		r.Body.Close()
		return fmt.Sprint(m["estado"])
	}
	if got := get(); got != "alquilado" {
		t.Fatalf("esperaba el inmueble 'alquilado' tras activar el contrato, obtuve %q", got)
	}

	// Rescindir -> vuelve a 'disponible'.
	upd := contratoBody(inmueble["id"], inquilino["id"])
	upd["fechaFin"] = body["fechaFin"]
	upd["estado"] = "rescindido"
	upd["motivoBaja"] = "acuerdo entre las partes"
	b, _ := json.Marshal(upd)
	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/contratos/%v", srv.URL, creado["id"]), bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	putResp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PUT: %v", err)
	}
	putResp.Body.Close()
	if putResp.StatusCode != http.StatusOK {
		t.Fatalf("esperaba 200 al rescindir, obtuve %d", putResp.StatusCode)
	}
	if got := get(); got != "disponible" {
		t.Fatalf("esperaba el inmueble 'disponible' tras rescindir, obtuve %q", got)
	}
}

func TestAPI_CrearContratoInmuebleInexistente(t *testing.T) {
	srv, _ := newTestServer(t)
	_, inquilino := postInquilino(t, srv, map[string]any{"nombreCompleto": "Ana", "documentoIdentidad": "1"})

	resp, _ := postContrato(t, srv, contratoBody(9999, inquilino["id"]))
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("esperaba 400 para un inmueble inexistente, obtuve %d", resp.StatusCode)
	}
}

func TestAPI_CrearContratoSinInquilinos(t *testing.T) {
	srv, _ := newTestServer(t)
	_, inmueble := postInmueble(t, srv, map[string]any{"nombre": "x", "direccion": "y", "tipo": "piso"})

	body := contratoBody(inmueble["id"], 0)
	body["inquilinoIds"] = []any{}
	resp, _ := postContrato(t, srv, body)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("esperaba 400 sin co-arrendatarios, obtuve %d", resp.StatusCode)
	}
}

func TestAPI_GetContratoInexistente(t *testing.T) {
	srv, _ := newTestServer(t)
	resp, err := http.Get(srv.URL + "/api/contratos/999")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("esperaba 404, obtuve %d", resp.StatusCode)
	}
}

func TestAPI_HistoricoInquilinoVacio(t *testing.T) {
	srv, _ := newTestServer(t)
	_, inquilino := postInquilino(t, srv, map[string]any{"nombreCompleto": "Ana", "documentoIdentidad": "1"})

	r, err := http.Get(fmt.Sprintf("%s/api/inquilinos/%v/contratos", srv.URL, inquilino["id"]))
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		t.Fatalf("esperaba 200, obtuve %d", r.StatusCode)
	}
	var hist []map[string]any
	json.NewDecoder(r.Body).Decode(&hist)
	if len(hist) != 0 {
		t.Fatalf("esperaba histórico vacío, obtuve %+v", hist)
	}
}
