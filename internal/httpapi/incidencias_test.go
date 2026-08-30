package httpapi_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"
)

func postIncidencia(t *testing.T, srv *httptest.Server, inmuebleID any, body map[string]any) (*http.Response, map[string]any) {
	t.Helper()
	b, _ := json.Marshal(body)
	resp, err := http.Post(fmt.Sprintf("%s/api/inmuebles/%v/incidencias", srv.URL, inmuebleID), "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("POST incidencia: %v", err)
	}
	defer resp.Body.Close()
	var out map[string]any
	json.NewDecoder(resp.Body).Decode(&out)
	return resp, out
}

func putIncidencia(t *testing.T, srv *httptest.Server, id any, body map[string]any) (*http.Response, map[string]any) {
	t.Helper()
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/incidencias/%v", srv.URL, id), bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PUT incidencia: %v", err)
	}
	defer resp.Body.Close()
	var out map[string]any
	json.NewDecoder(resp.Body).Decode(&out)
	return resp, out
}

// incidenciaBody es el cuerpo mínimo (los dos campos obligatorios) más lo
// que se le añada por caso.
func incidenciaBody(extra map[string]any) map[string]any {
	body := map[string]any{
		"titulo":    "Fuga en el grifo de la cocina",
		"categoria": "fontaneria",
		"prioridad": "alta",
	}
	for k, v := range extra {
		body[k] = v
	}
	return body
}

func crearInmuebleRapido(t *testing.T, srv *httptest.Server) map[string]any {
	t.Helper()
	_, inmueble := postInmueble(t, srv, map[string]any{"nombre": "Bravo Murillo 210", "direccion": "Calle Bravo Murillo 210", "tipo": "piso"})
	return inmueble
}

func TestAPI_CrearIncidenciaApareceEnListadoYContador(t *testing.T) {
	srv, _ := newTestServer(t)
	inmueble := crearInmuebleRapido(t, srv)

	resp, creada := postIncidencia(t, srv, inmueble["id"], incidenciaBody(nil))
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("esperaba 201, obtuve %d: %+v", resp.StatusCode, creada)
	}
	if creada["estado"] != "abierta" {
		t.Fatalf("una incidencia nueva debería nacer 'abierta', obtuve %v", creada["estado"])
	}
	if creada["fechaApertura"] == nil || creada["fechaApertura"] == "" {
		t.Fatalf("esperaba fechaApertura fijada, obtuve %v", creada["fechaApertura"])
	}

	r, _ := http.Get(fmt.Sprintf("%s/api/inmuebles/%v/incidencias", srv.URL, inmueble["id"]))
	var lista []map[string]any
	json.NewDecoder(r.Body).Decode(&lista)
	r.Body.Close()
	if len(lista) != 1 {
		t.Fatalf("esperaba 1 incidencia en el listado del inmueble, obtuve %+v", lista)
	}
	abiertas := 0
	for _, inc := range lista {
		if inc["estado"] != "cerrada" {
			abiertas++
		}
	}
	if abiertas != 1 {
		t.Fatalf("el contador del tab (incidencias no cerradas) debería ser 1, obtuve %d", abiertas)
	}
}

func TestAPI_IncidenciaFlujoDeEstadoCompletoYFechado(t *testing.T) {
	srv, _ := newTestServer(t)
	inmueble := crearInmuebleRapido(t, srv)
	_, creada := postIncidencia(t, srv, inmueble["id"], incidenciaBody(nil))

	for _, estado := range []string{"en_proceso", "esperando_proveedor", "resuelta", "cerrada"} {
		resp, out := putIncidencia(t, srv, creada["id"], incidenciaBody(map[string]any{"estado": estado}))
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("esperaba 200 al pasar a %q, obtuve %d: %+v", estado, resp.StatusCode, out)
		}
		if out["estado"] != estado {
			t.Fatalf("esperaba estado %q, obtuve %v", estado, out["estado"])
		}
	}

	// Releer y comprobar que cada cambio quedó registrado con su fecha.
	r, _ := http.Get(fmt.Sprintf("%s/api/incidencias/%v", srv.URL, creada["id"]))
	var inc map[string]any
	json.NewDecoder(r.Body).Decode(&inc)
	r.Body.Close()

	eventos, _ := inc["eventos"].([]any)
	cambios := 0
	for _, raw := range eventos {
		e := raw.(map[string]any)
		if e["creadoEn"] == nil || e["creadoEn"] == "" {
			t.Fatalf("todos los eventos deben ir fechados, este no: %+v", e)
		}
		if e["tipo"] == "cambio_estado" {
			cambios++
		}
	}
	if cambios != 4 {
		t.Fatalf("esperaba 4 cambios de estado registrados (uno por paso del flujo), obtuve %d: %+v", cambios, eventos)
	}
	if inc["fechaCierre"] == nil || inc["fechaCierre"] == "" {
		t.Fatalf("una incidencia cerrada debería llevar fechaCierre, obtuve %v", inc["fechaCierre"])
	}
}

func TestAPI_IncidenciaTransicionQueSaltaPasosSeRechaza(t *testing.T) {
	srv, _ := newTestServer(t)
	inmueble := crearInmuebleRapido(t, srv)
	_, creada := postIncidencia(t, srv, inmueble["id"], incidenciaBody(nil))

	// abierta -> resuelta se salta "en proceso" y "esperando proveedor".
	resp, out := putIncidencia(t, srv, creada["id"], incidenciaBody(map[string]any{"estado": "resuelta"}))
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("esperaba 409 al saltarse pasos del flujo, obtuve %d: %+v", resp.StatusCode, out)
	}
}

func TestAPI_IncidenciaEstadoFueraDelCheckDaError(t *testing.T) {
	srv, _ := newTestServer(t)
	inmueble := crearInmuebleRapido(t, srv)
	_, creada := postIncidencia(t, srv, inmueble["id"], incidenciaBody(nil))

	resp, _ := putIncidencia(t, srv, creada["id"], incidenciaBody(map[string]any{"estado": "pendiente"}))
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("esperaba 400 para un estado fuera de los permitidos, obtuve %d", resp.StatusCode)
	}
}

func TestAPI_CrearIncidenciaSinTitulo(t *testing.T) {
	srv, _ := newTestServer(t)
	inmueble := crearInmuebleRapido(t, srv)

	body := incidenciaBody(nil)
	delete(body, "titulo")
	resp, _ := postIncidencia(t, srv, inmueble["id"], body)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("esperaba 400 sin el campo obligatorio titulo, obtuve %d", resp.StatusCode)
	}
}

func TestAPI_CrearIncidenciaPrioridadFueraDelCheck(t *testing.T) {
	srv, _ := newTestServer(t)
	inmueble := crearInmuebleRapido(t, srv)

	resp, _ := postIncidencia(t, srv, inmueble["id"], incidenciaBody(map[string]any{"prioridad": "altisima"}))
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("esperaba 400 para una prioridad fuera de los permitidos, obtuve %d", resp.StatusCode)
	}
}

func TestAPI_IncidenciaInmuebleInexistente(t *testing.T) {
	srv, _ := newTestServer(t)
	resp, _ := postIncidencia(t, srv, 9999, incidenciaBody(nil))
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("esperaba 404 para un inmueble inexistente, obtuve %d", resp.StatusCode)
	}
}

func TestAPI_GetIncidenciaInexistente(t *testing.T) {
	srv, _ := newTestServer(t)
	resp, err := http.Get(srv.URL + "/api/incidencias/999")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("esperaba 404, obtuve %d", resp.StatusCode)
	}
}

func TestAPI_IncidenciaCosteACargoDeSeDistingue(t *testing.T) {
	srv, _ := newTestServer(t)
	inmueble := crearInmuebleRapido(t, srv)

	_, prop := postIncidencia(t, srv, inmueble["id"], incidenciaBody(map[string]any{
		"titulo": "Caldera", "coste": 85, "costeACargoDe": "propietario",
	}))
	_, inq := postIncidencia(t, srv, inmueble["id"], incidenciaBody(map[string]any{
		"titulo": "Cristal roto", "coste": 40, "costeACargoDe": "inquilino",
	}))

	if prop["costeACargoDe"] != "propietario" || prop["coste"].(float64) != 85 {
		t.Fatalf("esperaba 85 a cargo del propietario, obtuve %v / %v", prop["coste"], prop["costeACargoDe"])
	}
	if inq["costeACargoDe"] != "inquilino" || inq["coste"].(float64) != 40 {
		t.Fatalf("esperaba 40 a cargo del inquilino, obtuve %v / %v", inq["coste"], inq["costeACargoDe"])
	}

	resp, _ := postIncidencia(t, srv, inmueble["id"], incidenciaBody(map[string]any{"costeACargoDe": "vecino"}))
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("esperaba 400 para un 'a cargo de' fuera de los permitidos, obtuve %d", resp.StatusCode)
	}
}

func TestAPI_IncidenciaFotosAdjuntas(t *testing.T) {
	srv, _ := newTestServer(t)
	inmueble := crearInmuebleRapido(t, srv)
	_, creada := postIncidencia(t, srv, inmueble["id"], incidenciaBody(nil))

	contenido := []byte{0xFF, 0xD8, 0xFF, 0xE0, 'f', 'u', 'g', 'a'}
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	header := textproto.MIMEHeader{}
	header.Set("Content-Disposition", `form-data; name="archivo"; filename="fuga.jpg"`)
	header.Set("Content-Type", "image/jpeg")
	part, _ := mw.CreatePart(header)
	part.Write(contenido)
	mw.Close()

	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/incidencias/%v/documentos", srv.URL, creada["id"]), &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	up, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST documento incidencia: %v", err)
	}
	defer up.Body.Close()
	if up.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(up.Body)
		t.Fatalf("esperaba 201 al subir la foto, obtuve %d: %s", up.StatusCode, b)
	}
	var doc map[string]any
	json.NewDecoder(up.Body).Decode(&doc)
	if doc["entidadTipo"] != "incidencia" {
		t.Fatalf("el documento debería quedar asociado a la entidad 'incidencia', obtuve %v", doc["entidadTipo"])
	}

	// Se lista bajo la incidencia y se descarga byte a byte igual que el original.
	r, _ := http.Get(fmt.Sprintf("%s/api/incidencias/%v/documentos", srv.URL, creada["id"]))
	var docs []map[string]any
	json.NewDecoder(r.Body).Decode(&docs)
	r.Body.Close()
	if len(docs) != 1 {
		t.Fatalf("esperaba 1 foto listada bajo la incidencia, obtuve %+v", docs)
	}

	dl, _ := http.Get(fmt.Sprintf("%s/api/documentos/%v", srv.URL, doc["id"]))
	recibido, _ := io.ReadAll(dl.Body)
	dl.Body.Close()
	if !bytes.Equal(recibido, contenido) {
		t.Fatalf("el contenido descargado no coincide con el subido: %v != %v", recibido, contenido)
	}
}

func TestAPI_IncidenciaComentarioSeGuardaEnElHistorial(t *testing.T) {
	srv, _ := newTestServer(t)
	inmueble := crearInmuebleRapido(t, srv)
	_, creada := postIncidencia(t, srv, inmueble["id"], incidenciaBody(nil))

	resp, _ := putIncidencia(t, srv, creada["id"], incidenciaBody(map[string]any{
		"estado":     "en_proceso",
		"comentario": "Llamado el fontanero, viene el jueves",
	}))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("esperaba 200, obtuve %d", resp.StatusCode)
	}

	r, _ := http.Get(fmt.Sprintf("%s/api/incidencias/%v", srv.URL, creada["id"]))
	var inc map[string]any
	json.NewDecoder(r.Body).Decode(&inc)
	r.Body.Close()

	eventos, _ := inc["eventos"].([]any)
	tieneComentario := false
	for _, raw := range eventos {
		e := raw.(map[string]any)
		if e["tipo"] == "comentario" && e["comentario"] == "Llamado el fontanero, viene el jueves" {
			tieneComentario = true
		}
	}
	if !tieneComentario {
		t.Fatalf("esperaba el comentario en el historial, eventos = %+v", eventos)
	}
	// El comentario no se persiste en la propia fila de la incidencia.
	if _, ok := inc["comentario"]; ok {
		t.Fatalf("el campo comentario no debería devolverse en la incidencia, obtuve %v", inc["comentario"])
	}
}
