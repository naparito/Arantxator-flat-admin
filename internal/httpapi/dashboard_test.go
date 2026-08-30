package httpapi_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func isoEnDias(dias int) string {
	return time.Now().AddDate(0, 0, dias).Format("2006-01-02")
}

// contratoBodyFechas parte del cuerpo mínimo y sustituye firma/inicio/fin.
func contratoBodyFechas(inmuebleID, inquilinoID any, firma, inicio, fin string) map[string]any {
	b := contratoBody(inmuebleID, inquilinoID)
	b["fechaFirma"] = firma
	b["fechaInicio"] = inicio
	b["fechaFin"] = fin
	return b
}

type notifResp struct {
	Notificaciones []struct {
		Clave       string `json:"clave"`
		Tipo        string `json:"tipo"`
		Severidad   string `json:"severidad"`
		EntidadID   int64  `json:"entidadId"`
		Leida       bool   `json:"leida"`
		Descripcion string `json:"descripcion"`
	} `json:"notificaciones"`
	TotalActivas int `json:"totalActivas"`
	TotalSinLeer int `json:"totalSinLeer"`
}

func getNotificaciones(t *testing.T, srv *httptest.Server) notifResp {
	t.Helper()
	var out notifResp
	resp := getInto(t, srv, "/api/notificaciones", &out)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /api/notificaciones: esperaba 200, obtuve %d", resp.StatusCode)
	}
	return out
}

func TestAPI_Notificaciones_ContratoDentroDeLaVentanaGeneraSuAviso(t *testing.T) {
	srv, _ := newTestServer(t)
	inmueble, inquilino := crearInmuebleYInquilino(t, srv, false)

	// Contrato que vence dentro de 40 días (< ventana de aviso de 60) y con
	// la fianza recién firmada (aún pendiente).
	resp, creado := postContrato(t, srv, contratoBodyFechas(
		inmueble["id"], inquilino["id"], isoEnDias(0), isoEnDias(0), isoEnDias(40)))
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("esperaba 201 al crear el contrato, obtuve %d: %+v", resp.StatusCode, creado)
	}
	contratoID := int64(creado["id"].(float64))

	n := getNotificaciones(t, srv)
	var vioVencimiento, vioFianza bool
	for _, x := range n.Notificaciones {
		if x.Tipo == "contrato_por_vencer" && x.EntidadID == contratoID {
			vioVencimiento = true
			if x.Severidad != "aviso" {
				t.Errorf("un contrato a 40 días debería ser severidad 'aviso', es %q", x.Severidad)
			}
		}
		if x.Tipo == "fianza_sin_depositar" && x.EntidadID == contratoID {
			vioFianza = true
		}
	}
	if !vioVencimiento {
		t.Fatalf("esperaba un aviso 'contrato_por_vencer' para el contrato %d, avisos: %+v", contratoID, n.Notificaciones)
	}
	if !vioFianza {
		t.Fatalf("esperaba también un aviso 'fianza_sin_depositar' para el contrato %d", contratoID)
	}
	if n.TotalActivas < 2 || n.TotalSinLeer != n.TotalActivas {
		t.Fatalf("esperaba todos los avisos sin leer (activas=%d sinLeer=%d)", n.TotalActivas, n.TotalSinLeer)
	}
}

func TestAPI_Notificaciones_FianzaUrgenteVsAviso(t *testing.T) {
	srv, _ := newTestServer(t)

	// Inmueble A: firma hace 27 días -> límite de la fianza dentro de 3 días -> urgente.
	_, inmA := postInmueble(t, srv, map[string]any{"nombre": "A", "direccion": "Calle A 1", "tipo": "piso"})
	_, inqA := postInquilino(t, srv, map[string]any{"nombreCompleto": "Ana", "documentoIdentidad": "1"})
	respA, cA := postContrato(t, srv, contratoBodyFechas(inmA["id"], inqA["id"], isoEnDias(-27), isoEnDias(-27), isoEnDias(1800)))
	if respA.StatusCode != http.StatusCreated {
		t.Fatalf("contrato A: esperaba 201, obtuve %d: %+v", respA.StatusCode, cA)
	}

	// Inmueble B: firma hoy -> límite dentro de 30 días -> aviso (hay margen).
	_, inmB := postInmueble(t, srv, map[string]any{"nombre": "B", "direccion": "Calle B 2", "tipo": "piso"})
	_, inqB := postInquilino(t, srv, map[string]any{"nombreCompleto": "Bruno", "documentoIdentidad": "2"})
	respB, cB := postContrato(t, srv, contratoBodyFechas(inmB["id"], inqB["id"], isoEnDias(0), isoEnDias(0), isoEnDias(1800)))
	if respB.StatusCode != http.StatusCreated {
		t.Fatalf("contrato B: esperaba 201, obtuve %d: %+v", respB.StatusCode, cB)
	}
	idA := int64(cA["id"].(float64))
	idB := int64(cB["id"].(float64))

	sev := map[int64]string{}
	for _, x := range getNotificaciones(t, srv).Notificaciones {
		if x.Tipo == "fianza_sin_depositar" {
			sev[x.EntidadID] = x.Severidad
		}
	}
	if sev[idA] != "urgente" {
		t.Fatalf("fianza con 3 días de plazo debería ser 'urgente', es %q", sev[idA])
	}
	if sev[idB] != "aviso" {
		t.Fatalf("fianza con 30 días de margen debería ser 'aviso', es %q", sev[idB])
	}
}

func TestAPI_Notificaciones_MarcarLeidaRetiraDelContadorPeroNoTocaElDato(t *testing.T) {
	srv, _ := newTestServer(t)
	inmueble, inquilino := crearInmuebleYInquilino(t, srv, false)
	_, creado := postContrato(t, srv, contratoBodyFechas(
		inmueble["id"], inquilino["id"], isoEnDias(0), isoEnDias(0), isoEnDias(40)))
	contratoID := int64(creado["id"].(float64))

	antes := getNotificaciones(t, srv)
	var clave string
	for _, x := range antes.Notificaciones {
		if x.Tipo == "contrato_por_vencer" && x.EntidadID == contratoID {
			clave = x.Clave
		}
	}
	if clave == "" {
		t.Fatalf("no encontré el aviso de vencimiento a marcar; avisos: %+v", antes.Notificaciones)
	}

	req, _ := http.NewRequest(http.MethodPost,
		fmt.Sprintf("%s/api/notificaciones/%s/leida", srv.URL, url.PathEscape(clave)), nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST leida: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("esperaba 200 al marcar leída, obtuve %d", resp.StatusCode)
	}

	despues := getNotificaciones(t, srv)
	if despues.TotalActivas != antes.TotalActivas {
		t.Fatalf("marcar leída no debería cambiar el nº de avisos activos (%d -> %d)", antes.TotalActivas, despues.TotalActivas)
	}
	if despues.TotalSinLeer != antes.TotalSinLeer-1 {
		t.Fatalf("marcar leída debería bajar el contador de sin leer en 1 (%d -> %d)", antes.TotalSinLeer, despues.TotalSinLeer)
	}
	for _, x := range despues.Notificaciones {
		if x.Clave == clave && !x.Leida {
			t.Fatalf("el aviso %q debería figurar como leído", clave)
		}
	}

	// El dato subyacente no se ha tocado: el contrato sigue "próximo a vencer".
	var contrato map[string]any
	getInto(t, srv, fmt.Sprintf("/api/contratos/%d", contratoID), &contrato)
	if contrato["estado"] != "proximo_a_vencer" {
		t.Fatalf("el contrato debería seguir 'proximo_a_vencer' aunque el aviso esté leído, es %v", contrato["estado"])
	}
}

func TestAPI_Notificaciones_ClaveInvalida(t *testing.T) {
	srv, _ := newTestServer(t)
	for _, clave := range []string{"basura", "contrato_por_vencer:contrato:abc", "inventado:contrato:1", "contrato_por_vencer:inquilino:1"} {
		req, _ := http.NewRequest(http.MethodPost,
			fmt.Sprintf("%s/api/notificaciones/%s/leida", srv.URL, url.PathEscape(clave)), nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("POST leida %q: %v", clave, err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("clave %q: esperaba 400, obtuve %d", clave, resp.StatusCode)
		}
	}
}

func TestAPI_DashboardResumen_SeRecalculaSobreDatosReales(t *testing.T) {
	srv, _ := newTestServer(t)
	inmueble, inquilino := crearInmuebleYInquilino(t, srv, false)
	inmuebleID := inmueble["id"]

	// Contrato próximo a vencer (40 días).
	postContrato(t, srv, contratoBodyFechas(inmuebleID, inquilino["id"], isoEnDias(0), isoEnDias(0), isoEnDias(40)))

	mes := time.Now().Format("2006-01")
	// Cobro de renta del mes en curso: 980 €.
	postJSON(t, srv, fmt.Sprintf("/api/inmuebles/%v/cobros", inmuebleID), map[string]any{
		"periodo": mes + "-01", "importe": 980, "metodoPago": "transferencia",
	})
	// Factura pendiente de 113 € emitida este mes.
	postJSON(t, srv, fmt.Sprintf("/api/inmuebles/%v/gastos", inmuebleID), map[string]any{
		"tipo": "comunidad", "importe": 113, "fechaEmision": mes + "-15", "fechaVencimiento": isoEnDias(10),
	})

	var resumen struct {
		Ocupacion struct {
			InmueblesTotales  int `json:"inmueblesTotales"`
			InmueblesOcupados int `json:"inmueblesOcupados"`
			Porcentaje        int `json:"porcentaje"`
		} `json:"ocupacion"`
		ContratosPorVencer int `json:"contratosPorVencer"`
		GastosPendientes   struct {
			Cantidad int     `json:"cantidad"`
			Importe  float64 `json:"importe"`
		} `json:"gastosPendientes"`
		Rentabilidad struct {
			Ingresos float64 `json:"ingresos"`
			Gastos   float64 `json:"gastos"`
			Neto     float64 `json:"neto"`
		} `json:"rentabilidad"`
	}
	getInto(t, srv, "/api/dashboard/resumen", &resumen)

	if resumen.ContratosPorVencer != 1 {
		t.Fatalf("esperaba 1 contrato por vencer, obtuve %d", resumen.ContratosPorVencer)
	}
	if resumen.Ocupacion.InmueblesTotales != 1 || resumen.Ocupacion.InmueblesOcupados != 1 || resumen.Ocupacion.Porcentaje != 100 {
		t.Fatalf("ocupación inesperada: %+v", resumen.Ocupacion)
	}
	if resumen.GastosPendientes.Cantidad != 1 || resumen.GastosPendientes.Importe != 113 {
		t.Fatalf("gastos pendientes inesperados: %+v", resumen.GastosPendientes)
	}
	if resumen.Rentabilidad.Ingresos != 980 || resumen.Rentabilidad.Gastos != 113 || resumen.Rentabilidad.Neto != 867 {
		t.Fatalf("rentabilidad agregada inesperada: %+v", resumen.Rentabilidad)
	}

	// Añadir otra factura pendiente y volver a pedir el resumen: los números
	// se recalculan (no vienen de una cifra cacheada).
	postJSON(t, srv, fmt.Sprintf("/api/inmuebles/%v/gastos", inmuebleID), map[string]any{
		"tipo": "agua", "importe": 50, "fechaEmision": mes + "-16", "fechaVencimiento": isoEnDias(12),
	})
	getInto(t, srv, "/api/dashboard/resumen", &resumen)
	if resumen.GastosPendientes.Cantidad != 2 || resumen.GastosPendientes.Importe != 163 {
		t.Fatalf("tras añadir una factura esperaba 2 pendientes por 163 €, obtuve %+v", resumen.GastosPendientes)
	}
	if resumen.Rentabilidad.Gastos != 163 || resumen.Rentabilidad.Neto != 817 {
		t.Fatalf("la rentabilidad debería reflejar el nuevo gasto: %+v", resumen.Rentabilidad)
	}
}
