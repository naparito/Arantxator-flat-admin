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

func postJSON(t *testing.T, srv *httptest.Server, path string, body map[string]any) (*http.Response, map[string]any) {
	t.Helper()
	b, _ := json.Marshal(body)
	resp, err := http.Post(srv.URL+path, "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("POST %s: %v", path, err)
	}
	defer resp.Body.Close()
	var out map[string]any
	json.NewDecoder(resp.Body).Decode(&out)
	return resp, out
}

func putJSON(t *testing.T, srv *httptest.Server, path string, body map[string]any) (*http.Response, map[string]any) {
	t.Helper()
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPut, srv.URL+path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PUT %s: %v", path, err)
	}
	defer resp.Body.Close()
	var out map[string]any
	json.NewDecoder(resp.Body).Decode(&out)
	return resp, out
}

func getInto(t *testing.T, srv *httptest.Server, path string, dst any) *http.Response {
	t.Helper()
	resp, err := http.Get(srv.URL + path)
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	defer resp.Body.Close()
	if dst != nil {
		json.NewDecoder(resp.Body).Decode(dst)
	}
	return resp
}

// pisoCompartidoConTresInquilinos crea un inmueble compartido y tres
// inquilinos, y devuelve el id del inmueble y los tres ids.
func pisoCompartidoConTresInquilinos(t *testing.T, srv *httptest.Server) (float64, [3]float64) {
	t.Helper()
	_, inmueble := postInmueble(t, srv, map[string]any{
		"nombre": "Bravo Murillo 210", "direccion": "Calle Bravo Murillo 210, Bajo A", "tipo": "piso", "compartido": true,
	})
	var ids [3]float64
	for i, nombre := range []string{"Javier Martín Soto", "Ana Belén Torres", "Pablo Navarro Castillo"} {
		_, inq := postInquilino(t, srv, map[string]any{"nombreCompleto": nombre, "documentoIdentidad": fmt.Sprintf("D%d", i)})
		ids[i] = inq["id"].(float64)
	}
	return inmueble["id"].(float64), ids
}

func gastoLuzBody(fechaEmision string) map[string]any {
	return map[string]any{
		"tipo":             "luz",
		"periodicidad":     "mensual",
		"importe":          78.00,
		"fechaEmision":     fechaEmision,
		"fechaVencimiento": "2026-09-30",
		"proveedor":        "Iberdrola",
	}
}

func TestAPI_Reparto333334_FacturaLuz_ReciboCuadra(t *testing.T) {
	srv, _ := newTestServer(t)
	inmuebleID, inq := pisoCompartidoConTresInquilinos(t, srv)

	// Reparto 33/33/34 para "luz", vigente desde 2026-01-01.
	resp, _ := postJSON(t, srv, fmt.Sprintf("/api/inmuebles/%v/reparto", inmuebleID), map[string]any{
		"tipoGasto":    "luz",
		"vigenteDesde": "2026-01-01",
		"motivo":       "reparto inicial",
		"cuotas": []map[string]any{
			{"inquilinoId": inq[0], "porcentaje": 33},
			{"inquilinoId": inq[1], "porcentaje": 33},
			{"inquilinoId": inq[2], "porcentaje": 34},
		},
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("esperaba 201 al crear el reparto, obtuve %d", resp.StatusCode)
	}

	// Alta de la factura de luz de 78,00 €.
	resp, gasto := postJSON(t, srv, fmt.Sprintf("/api/inmuebles/%v/gastos", inmuebleID), gastoLuzBody("2026-09-01"))
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("esperaba 201 al crear el gasto, obtuve %d: %+v", resp.StatusCode, gasto)
	}

	var recibo struct {
		Total      float64 `json:"total"`
		SinReparto bool    `json:"sinReparto"`
		Lineas     []struct {
			InquilinoID float64 `json:"inquilinoId"`
			Porcentaje  float64 `json:"porcentaje"`
			Importe     float64 `json:"importe"`
		} `json:"lineas"`
	}
	getInto(t, srv, fmt.Sprintf("/api/gastos/%v/recibo", gasto["id"]), &recibo)

	if recibo.SinReparto || len(recibo.Lineas) != 3 {
		t.Fatalf("esperaba un recibo con 3 líneas, obtuve %+v", recibo)
	}
	quiero := map[float64]float64{inq[0]: 25.74, inq[1]: 25.74, inq[2]: 26.52}
	suma := 0.0
	for _, l := range recibo.Lineas {
		if l.Importe != quiero[l.InquilinoID] {
			t.Fatalf("inquilino %v: esperaba %.2f €, obtuve %.2f €", l.InquilinoID, quiero[l.InquilinoID], l.Importe)
		}
		suma += l.Importe
	}
	if fmt.Sprintf("%.2f", suma) != "78.00" {
		t.Fatalf("la suma de los recibos (%.2f) no cuadra con el total 78,00 €", suma)
	}
}

func TestAPI_RepartoQueNoSuma100EsRechazado(t *testing.T) {
	srv, _ := newTestServer(t)
	inmuebleID, inq := pisoCompartidoConTresInquilinos(t, srv)

	resp, out := postJSON(t, srv, fmt.Sprintf("/api/inmuebles/%v/reparto", inmuebleID), map[string]any{
		"tipoGasto":    "luz",
		"vigenteDesde": "2026-01-01",
		"cuotas": []map[string]any{
			{"inquilinoId": inq[0], "porcentaje": 33},
			{"inquilinoId": inq[1], "porcentaje": 33},
			{"inquilinoId": inq[2], "porcentaje": 30}, // suma 96
		},
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("esperaba 400 para un reparto que no suma 100, obtuve %d: %+v", resp.StatusCode, out)
	}

	// Y no se ha guardado nada a medias.
	var got struct {
		Versiones []any `json:"versiones"`
	}
	getInto(t, srv, fmt.Sprintf("/api/inmuebles/%v/reparto", inmuebleID), &got)
	if len(got.Versiones) != 0 {
		t.Fatalf("un reparto inválido no debería dejar filas, obtuve %+v", got.Versiones)
	}
}

func TestAPI_FacturaAnteriorAlCambioDeRepartoUsaElRepartoDeEntonces(t *testing.T) {
	srv, _ := newTestServer(t)
	inmuebleID, inq := pisoCompartidoConTresInquilinos(t, srv)

	// v1: dos inquilinos al 50 %, vigente desde 2026-01-01.
	postJSON(t, srv, fmt.Sprintf("/api/inmuebles/%v/reparto", inmuebleID), map[string]any{
		"tipoGasto": "luz", "vigenteDesde": "2026-01-01", "motivo": "inicial",
		"cuotas": []map[string]any{
			{"inquilinoId": inq[0], "porcentaje": 50},
			{"inquilinoId": inq[1], "porcentaje": 50},
		},
	})
	// v2: entra el tercero, 33/33/34, vigente desde 2026-06-01.
	postJSON(t, srv, fmt.Sprintf("/api/inmuebles/%v/reparto", inmuebleID), map[string]any{
		"tipoGasto": "luz", "vigenteDesde": "2026-06-01", "motivo": "entrada de Pablo",
		"cuotas": []map[string]any{
			{"inquilinoId": inq[0], "porcentaje": 33},
			{"inquilinoId": inq[1], "porcentaje": 33},
			{"inquilinoId": inq[2], "porcentaje": 34},
		},
	})

	// Factura con fecha de emisión ANTERIOR al cambio (marzo).
	_, gastoViejo := postJSON(t, srv, fmt.Sprintf("/api/inmuebles/%v/gastos", inmuebleID), gastoLuzBody("2026-03-15"))
	var reciboViejo struct {
		Lineas []struct {
			Importe float64 `json:"importe"`
		} `json:"lineas"`
	}
	getInto(t, srv, fmt.Sprintf("/api/gastos/%v/recibo", gastoViejo["id"]), &reciboViejo)
	if len(reciboViejo.Lineas) != 2 {
		t.Fatalf("una factura de marzo debería repartirse con la v1 (2 inquilinos), obtuve %+v", reciboViejo.Lineas)
	}
	if reciboViejo.Lineas[0].Importe != 39.00 || reciboViejo.Lineas[1].Importe != 39.00 {
		t.Fatalf("esperaba 39,00 € + 39,00 € (50/50 de 78), obtuve %+v", reciboViejo.Lineas)
	}

	// Factura posterior al cambio (septiembre): reparto nuevo, 3 inquilinos.
	_, gastoNuevo := postJSON(t, srv, fmt.Sprintf("/api/inmuebles/%v/gastos", inmuebleID), gastoLuzBody("2026-09-15"))
	var reciboNuevo struct {
		Lineas []struct {
			Importe float64 `json:"importe"`
		} `json:"lineas"`
	}
	getInto(t, srv, fmt.Sprintf("/api/gastos/%v/recibo", gastoNuevo["id"]), &reciboNuevo)
	if len(reciboNuevo.Lineas) != 3 {
		t.Fatalf("una factura de septiembre debería repartirse con la v2 (3 inquilinos), obtuve %+v", reciboNuevo.Lineas)
	}
}

func TestAPI_GastoEnInmuebleNoCompartidoSinRepartoNoDaError(t *testing.T) {
	srv, _ := newTestServer(t)
	_, inmueble := postInmueble(t, srv, map[string]any{
		"nombre": "Alcalá 145", "direccion": "Calle de Alcalá 145, 3ºB", "tipo": "piso",
	})

	resp, gasto := postJSON(t, srv, fmt.Sprintf("/api/inmuebles/%v/gastos", inmueble["id"]), gastoLuzBody("2026-09-01"))
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("esperaba 201, obtuve %d: %+v", resp.StatusCode, gasto)
	}

	var recibo struct {
		SinReparto bool  `json:"sinReparto"`
		Lineas     []any `json:"lineas"`
	}
	r := getInto(t, srv, fmt.Sprintf("/api/gastos/%v/recibo", gasto["id"]), &recibo)
	if r.StatusCode != http.StatusOK {
		t.Fatalf("un gasto sin reparto debería devolver 200, obtuve %d", r.StatusCode)
	}
	if !recibo.SinReparto || len(recibo.Lineas) != 0 {
		t.Fatalf("esperaba un recibo sin reparto y sin líneas, obtuve %+v", recibo)
	}
}

func TestAPI_RentabilidadCoincideConElCalculoManual(t *testing.T) {
	srv, _ := newTestServer(t)
	_, inmueble := postInmueble(t, srv, map[string]any{
		"nombre": "Alcalá 145", "direccion": "dir", "tipo": "piso",
	})
	id := inmueble["id"]

	// Gastos de septiembre: 78 + 42 + 35 + 32 = 187. Uno en octubre que NO cuenta.
	for _, g := range []map[string]any{
		{"tipo": "luz", "importe": 78, "fechaEmision": "2026-09-05"},
		{"tipo": "agua", "importe": 42, "fechaEmision": "2026-09-10"},
		{"tipo": "gas", "importe": 35, "fechaEmision": "2026-09-20"},
		{"tipo": "internet", "importe": 32, "fechaEmision": "2026-09-28"},
		{"tipo": "comunidad", "importe": 999, "fechaEmision": "2026-10-01"},
	} {
		if resp, out := postJSON(t, srv, fmt.Sprintf("/api/inmuebles/%v/gastos", id), g); resp.StatusCode != http.StatusCreated {
			t.Fatalf("creando gasto %+v: %d %+v", g, resp.StatusCode, out)
		}
	}
	// Ingresos de septiembre: 1350. Uno de octubre que NO cuenta.
	for _, c := range []map[string]any{
		{"periodo": "2026-09-01", "importe": 1350, "metodoPago": "transferencia"},
		{"periodo": "2026-10-01", "importe": 1350},
	} {
		if resp, out := postJSON(t, srv, fmt.Sprintf("/api/inmuebles/%v/cobros", id), c); resp.StatusCode != http.StatusCreated {
			t.Fatalf("creando cobro %+v: %d %+v", c, resp.StatusCode, out)
		}
	}

	var rent struct {
		Periodo  string  `json:"periodo"`
		Ingresos float64 `json:"ingresos"`
		Gastos   float64 `json:"gastos"`
		Neto     float64 `json:"neto"`
	}
	getInto(t, srv, fmt.Sprintf("/api/inmuebles/%v/rentabilidad?periodo=2026-09", id), &rent)

	if rent.Ingresos != 1350 || rent.Gastos != 187 || rent.Neto != 1163 {
		t.Fatalf("esperaba ingresos 1350, gastos 187, neto 1163 (cálculo manual); obtuve %+v", rent)
	}
}

func TestAPI_GastoTipoFueraDeLosPermitidos(t *testing.T) {
	srv, _ := newTestServer(t)
	_, inmueble := postInmueble(t, srv, map[string]any{"nombre": "x", "direccion": "y", "tipo": "piso"})

	body := gastoLuzBody("2026-09-01")
	body["tipo"] = "criptomonedas"
	resp, _ := postJSON(t, srv, fmt.Sprintf("/api/inmuebles/%v/gastos", inmueble["id"]), body)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("esperaba 400 para un tipo de gasto fuera de los permitidos, obtuve %d", resp.StatusCode)
	}
}

func TestAPI_GastoInexistente(t *testing.T) {
	srv, _ := newTestServer(t)
	if r := getInto(t, srv, "/api/gastos/9999", nil); r.StatusCode != http.StatusNotFound {
		t.Fatalf("esperaba 404 para un gasto inexistente, obtuve %d", r.StatusCode)
	}
	if r := getInto(t, srv, "/api/gastos/9999/recibo", nil); r.StatusCode != http.StatusNotFound {
		t.Fatalf("esperaba 404 para el recibo de un gasto inexistente, obtuve %d", r.StatusCode)
	}
}

func TestAPI_MarcarGastoPagado(t *testing.T) {
	srv, _ := newTestServer(t)
	_, inmueble := postInmueble(t, srv, map[string]any{"nombre": "x", "direccion": "y", "tipo": "piso"})
	_, gasto := postJSON(t, srv, fmt.Sprintf("/api/inmuebles/%v/gastos", inmueble["id"]), gastoLuzBody("2026-09-01"))

	body := gastoLuzBody("2026-09-01")
	body["estadoPago"] = "pagado"
	body["metodoPago"] = "domiciliado"
	resp, out := putJSON(t, srv, fmt.Sprintf("/api/gastos/%v", gasto["id"]), body)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("esperaba 200 al marcar pagada, obtuve %d: %+v", resp.StatusCode, out)
	}
	if out["estadoPago"] != "pagado" {
		t.Fatalf("esperaba estadoPago 'pagado', obtuve %v", out["estadoPago"])
	}
	if out["fechaPago"] == nil || out["fechaPago"] == "" {
		t.Fatalf("al marcar pagada debería quedar una fechaPago, obtuve %v", out["fechaPago"])
	}
}

func TestAPI_GastoPDFAdjunto(t *testing.T) {
	srv, _ := newTestServer(t)
	_, inmueble := postInmueble(t, srv, map[string]any{"nombre": "x", "direccion": "y", "tipo": "piso"})
	_, gasto := postJSON(t, srv, fmt.Sprintf("/api/inmuebles/%v/gastos", inmueble["id"]), gastoLuzBody("2026-09-01"))

	contenido := []byte("%PDF-1.4 factura luz")
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	header := textproto.MIMEHeader{}
	header.Set("Content-Disposition", `form-data; name="archivo"; filename="factura-luz.pdf"`)
	header.Set("Content-Type", "application/pdf")
	part, _ := mw.CreatePart(header)
	part.Write(contenido)
	mw.Close()

	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/gastos/%v/documentos", srv.URL, gasto["id"]), &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	up, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST documento gasto: %v", err)
	}
	defer up.Body.Close()
	if up.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(up.Body)
		t.Fatalf("esperaba 201 al subir el PDF, obtuve %d: %s", up.StatusCode, b)
	}
	var doc map[string]any
	json.NewDecoder(up.Body).Decode(&doc)
	if doc["entidadTipo"] != "gasto" {
		t.Fatalf("el documento debería quedar asociado a la entidad 'gasto', obtuve %v", doc["entidadTipo"])
	}

	var docs []map[string]any
	getInto(t, srv, fmt.Sprintf("/api/gastos/%v/documentos", gasto["id"]), &docs)
	if len(docs) != 1 {
		t.Fatalf("esperaba 1 documento listado bajo el gasto, obtuve %+v", docs)
	}

	dl, _ := http.Get(fmt.Sprintf("%s/api/documentos/%v", srv.URL, doc["id"]))
	recibido, _ := io.ReadAll(dl.Body)
	dl.Body.Close()
	if !bytes.Equal(recibido, contenido) {
		t.Fatalf("el contenido descargado no coincide con el subido")
	}
}
