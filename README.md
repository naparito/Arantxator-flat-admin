# Arantxator Flat Admin

Sistema de gestión integral de alquileres para **una** persona y **su** cartera:
inmuebles, inquilinos, contratos (con las reglas de la LAU para la Comunidad de
Madrid), incidencias, gastos con reparto porcentual entre inquilinos, y un panel
de resumen con centro de notificaciones. Todo en una aplicación de escritorio
autocontenida —un único `.exe` que abre el navegador—, local, sin cuentas ni
conexión a internet.

> **Estado:** **v1.0-alpha** — primera release (tag `v1.0-alpha`, rama `main`).
> Todos los módulos (hitos 0–7) están integrados y probados. `development` sigue
> siendo la rama de trabajo para lo que venga después. Pendiente sin bloquear la
> release: probar el instalador en una máquina Windows sin herramientas de
> desarrollo. Ver [`docs/plan/plan-implementacion.md`](docs/plan/plan-implementacion.md).

## Documentación

| Documento | Contenido |
|---|---|
| **[Manual funcional](docs/manual/README.md)** | **Cómo se usa cada módulo**, pantalla a pantalla, con sus reglas, diagramas y endpoints. Empieza por aquí. |
| [Diseño técnico y funcional](docs/design/diseno-tecnico-funcional.md) | Análisis de requisitos y decisiones de arquitectura (LAU, modelo de datos, alcance). |
| [Plan de implementación](docs/plan/plan-implementacion.md) | Hitos, entregables y batería de pruebas de cada uno. |
| [Instalación y despliegue](docs/despliegue/instalacion-despliegue.md) | Cómo se genera el instalador de Windows y cómo lo usa el usuario final. |
| [`web/README.md`](web/README.md) | Puesta en marcha de la SPA (Vite, tests, build embebido). |

### Los módulos de un vistazo

| Módulo | Qué resuelve | Manual |
|---|---|---|
| **Inmuebles** | Ficha de cada propiedad, documentación, suministros y —si es compartido— sus habitaciones. Incluye los submódulos Incidencias y Gastos. | [01](docs/manual/01-inmuebles.md) |
| **Inquilinos** | Ficha del inquilino, documentación (DNI, nómina, aval), IBAN e histórico de contratos. | [02](docs/manual/02-inquilinos.md) |
| **Contratos** | Arrendamientos con valores por defecto de la LAU (duración 5/7 años, fianza, plazo de depósito de 30 días, IRAV). El estado (`activo`/`próximo a vencer`/`vencido`) se **deriva** de las fechas. | [03](docs/manual/03-contratos.md) |
| **Incidencias** | Partes de mantenimiento por inmueble, con flujo de estados fechado (`abierta → … → cerrada`) y coste imputable. | [04](docs/manual/04-incidencias.md) |
| **Gastos y reparto** | Facturas por inmueble, reparto porcentual **versionado** por inquilino y tipo de gasto, recibo individual al céntimo, y rentabilidad neta (ingresos − gastos). | [05](docs/manual/05-gastos-y-reparto.md) |
| **Dashboard y notificaciones** | Resumen agregado de la cartera + centro de avisos evaluado en caliente (contrato por vencer, fianza sin depositar, factura pendiente, incidencia abierta), con aviso-resumen al arrancar. | [06](docs/manual/06-dashboard-y-notificaciones.md) |

Principio transversal: la aplicación guarda lo mínimo y **calcula el resto al
leer** (estado del contrato, % de ocupación, recibo individual, avisos, KPIs) —
nunca sobre una cifra cacheada. Detalle en el
[manual](docs/manual/README.md#conceptos-que-se-repiten-en-todos-los-módulos).

## Instalar (usuario final)

1. Descarga `Arantxator-Setup.exe` (se genera con `scripts/build.ps1`, ver abajo)
   y haz doble clic.
2. Sigue el asistente — sin conexión a internet ni permisos de administrador.
   Deja marcada la casilla para crear el acceso directo en el escritorio.
3. Al abrirlo, la aplicación arranca en segundo plano y el navegador se abre solo
   en `http://127.0.0.1:8080`. Si hay avisos pendientes, verás el aviso-resumen
   antes del panel.
4. **Copia de seguridad:** cerrar la app y copiar `arantxator.db` (junto al
   ejecutable) es la copia completa — datos y documentos incluidos, porque todo
   vive en ese único fichero SQLite.

Detalle de instalación, desinstalación y solución de problemas:
[`docs/despliegue/instalacion-despliegue.md`](docs/despliegue/instalacion-despliegue.md).

## Stack

- **Backend:** Go — compila a un único binario nativo, sin runtime que instalar.
- **Base de datos:** SQLite embebida (driver puro Go, sin CGO). Los adjuntos
  (contratos, facturas, fotos, DNI…) se guardan como BLOB dentro de la propia
  base de datos ⇒ una copia de seguridad es siempre un único fichero `.db`.
- **GUI:** SPA (React + Vite + TypeScript) compilada y embebida en el binario con
  `go:embed`. Sin librería de componentes: CSS plano con el sistema de tokens de
  los mockups aprobados.
- **API interna:** HTTP + JSON (`/api/*`), consumida solo por la propia SPA.

## Estructura del repositorio

```
cmd/arantxator/           Punto de entrada del binario (+ icono/versión embebidos)
internal/
  config/                 Configuración de arranque (puerto, ruta de la BD)
  domain/                 Entidades y reglas de negocio puras:
                            Inmueble, Habitacion, Inquilino, Contrato, Gasto,
                            RepartoGasto, CobroRenta, Incidencia, Documento,
                            Notificacion, ResumenDashboard (+ sus tests)
  httpapi/                Rutas y handlers de la API HTTP interna (+ tests httptest)
  storage/sqlite/         Conexión SQLite + migraciones embebidas (0001..0007)
                            + un repositorio por entidad (+ tests con BD temporal)
  webui/                  SPA compilada embebida, con fallback de rutas para que
                            React Router funcione al recargar o abrir un enlace directo
web/                      Código fuente de la SPA (React + Vite + TypeScript)
  src/api/                Cliente HTTP, tipos y contexto de notificaciones
  src/pages/              Una pantalla por fichero (Resumen, Notificaciones,
                            Inmuebles*, Inquilinos*, Contratos*, Gastos)
  src/components/         Layout, Sidebar, AvisoArranque, iconos
docs/manual/              Manual funcional por módulo  ← empieza aquí
docs/design/              Análisis de requisitos y diseño técnico/funcional
docs/plan/                Plan de implementación por hitos y su batería de pruebas
docs/despliegue/          Instalación y despliegue (build, instalador, icono)
installer/                Script de Inno Setup e icono de la aplicación
scripts/                  Scripts de compilación (SPA + binario + instalador)
```

## Compilar y ejecutar (desarrollo)

Requiere [Go](https://go.dev/dl/) 1.23+ y [Node.js](https://nodejs.org/) LTS.

```bash
git checkout development && git pull
go mod tidy
go run ./cmd/arantxator          # http://127.0.0.1:8080, abre el navegador solo
```

Sirve la SPA ya compilada que va versionada en `internal/webui/dist/`. Para
trabajar en la SPA con recarga en caliente, en otra terminal:

```bash
cd web && npm install && npm run dev   # Vite en :5173, proxy /api -> :8080
```

### Persistencia de datos con `go run`

`arantxator.db` se guarda **junto al ejecutable en marcha**. `go run` compila un
binario temporal en una carpeta distinta cada vez, así que "pierde" la base de
datos entre arranques. Para pruebas con persistencia real, fija la ruta:

```bash
export ARANTXATOR_DB_PATH="$HOME/arantxator-dev.db"   # PowerShell: $env:ARANTXATOR_DB_PATH = "..."
go run ./cmd/arantxator
```

| Variable de entorno | Por defecto | Uso |
|---|---|---|
| `ARANTXATOR_ADDR` | `127.0.0.1:8080` | Dirección y puerto del servidor. |
| `ARANTXATOR_DB_PATH` | `arantxator.db` junto al `.exe` | Ruta del fichero SQLite. |

Compilando el binario (`go build -o bin/arantxator.exe ./cmd/arantxator`) el
problema no existe: al vivir siempre en la misma ruta, `arantxator.db` persiste
sin fijar nada.

### Generar el binario final y el instalador de Windows

```powershell
powershell -File scripts/build.ps1
```

Encadena `npm run build` → `go build` → Inno Setup. Detalle (icono, versión,
verificación):
[`docs/despliegue/instalacion-despliegue.md`](docs/despliegue/instalacion-despliegue.md).

## Pruebas

Cada módulo se da por cerrado solo cuando pasa su batería (ver
[la estrategia en el plan](docs/plan/plan-implementacion.md#estrategia-de-pruebas)):

```bash
# Backend (Go)
go test ./...
go vet ./...

# Frontend (React)
cd web
npm test          # Vitest + React Testing Library
npm run build     # tsc -b + vite build (también valida tipos)
```

Además, antes de cerrar cada hito se repasa a mano la checklist de ese módulo y
la de los anteriores (no regresión) sobre el binario real (`bin/arantxator.exe`).

## Flujo de trabajo (git)

- `development` es la rama base de trabajo; `main` solo recibe versiones ya
  validadas.
- Cada hito/iteración se desarrolla en su propia rama (`feature/…`) creada desde
  `development` y se integra mediante pull request hacia `development`.

## Próximos pasos

- Llevar `development` a `main` como release de la **v1.0-alpha**.
- Prueba del instalador en una máquina Windows realmente limpia (sin Go/Node/Git)
  — única pregunta abierta del plan.
