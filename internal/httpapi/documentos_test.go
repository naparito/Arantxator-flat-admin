package httpapi_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"testing"
)

func subirDocumento(t *testing.T, baseURL string, inmuebleID any, nombreArchivo, tipoMime string, contenido []byte) *http.Response {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	header := textproto.MIMEHeader{}
	header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="archivo"; filename="%s"`, nombreArchivo))
	header.Set("Content-Type", tipoMime)
	part, err := mw.CreatePart(header)
	if err != nil {
		t.Fatalf("CreatePart: %v", err)
	}
	if _, err := part.Write(contenido); err != nil {
		t.Fatalf("escribiendo contenido: %v", err)
	}
	if err := mw.Close(); err != nil {
		t.Fatalf("cerrando multipart writer: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/inmuebles/%v/documentos", baseURL, inmuebleID), &buf)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST documento: %v", err)
	}
	return resp
}

func TestAPI_SubirYDescargarDocumento(t *testing.T) {
	srv, _ := newTestServer(t)

	_, inmueble := postInmueble(t, srv, map[string]any{"nombre": "x", "direccion": "y", "tipo": "piso"})
	inmuebleID := inmueble["id"]

	contenido := []byte{0xFF, 0xD8, 0xFF, 0xE0, 'f', 'o', 't', 'o'}
	resp := subirDocumento(t, srv.URL, inmuebleID, "fachada.jpg", "image/jpeg", contenido)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("esperaba 201 al subir el documento, obtuve %d: %s", resp.StatusCode, body)
	}
	var doc map[string]any
	json.NewDecoder(resp.Body).Decode(&doc)
	docID := doc["id"]
	if docID == nil {
		t.Fatalf("esperaba un id de documento en la respuesta: %+v", doc)
	}

	descarga, err := http.Get(fmt.Sprintf("%s/api/documentos/%v", srv.URL, docID))
	if err != nil {
		t.Fatalf("GET /api/documentos/{id}: %v", err)
	}
	defer descarga.Body.Close()
	if descarga.StatusCode != http.StatusOK {
		t.Fatalf("esperaba 200 al descargar, obtuve %d", descarga.StatusCode)
	}
	recibido, err := io.ReadAll(descarga.Body)
	if err != nil {
		t.Fatalf("leyendo cuerpo de la descarga: %v", err)
	}
	if !bytes.Equal(recibido, contenido) {
		t.Fatalf("el contenido descargado no es idéntico al subido: %v != %v", recibido, contenido)
	}
	if descarga.Header.Get("Content-Type") != "image/jpeg" {
		t.Fatalf("Content-Type inesperado: %s", descarga.Header.Get("Content-Type"))
	}

	// El documento debe aparecer en el listado del inmueble.
	listaResp, err := http.Get(fmt.Sprintf("%s/api/inmuebles/%v/documentos", srv.URL, inmuebleID))
	if err != nil {
		t.Fatalf("GET listado de documentos: %v", err)
	}
	defer listaResp.Body.Close()
	var lista []map[string]any
	json.NewDecoder(listaResp.Body).Decode(&lista)
	if len(lista) != 1 || lista[0]["nombreArchivo"] != "fachada.jpg" {
		t.Fatalf("esperaba el documento recién subido en el listado, obtuve %+v", lista)
	}
}

func TestAPI_DescargarDocumentoInexistente(t *testing.T) {
	srv, _ := newTestServer(t)

	resp, err := http.Get(srv.URL + "/api/documentos/999")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("esperaba 404, obtuve %d", resp.StatusCode)
	}
}

func TestAPI_SubirDocumentoAInmuebleInexistente(t *testing.T) {
	srv, _ := newTestServer(t)

	resp := subirDocumento(t, srv.URL, 999, "x.jpg", "image/jpeg", []byte("x"))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("esperaba 404 al subir a un inmueble inexistente, obtuve %d", resp.StatusCode)
	}
}
