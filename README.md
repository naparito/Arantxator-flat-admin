# Arantxator Flat Admin

Sistema de gestión integral de alquileres: inmuebles, inquilinos, contratos,
gastos con reparto porcentual e incidencias, en una aplicación visual,
autocontenida y pensada para un usuario sin conocimientos técnicos.

> **Estado:** módulo de Inmuebles funcionando de extremo a extremo (Hito 1,
> incluidos inmuebles compartidos por habitaciones), más el instalador de
> Windows ya listo (Hito 7, adelantado). Ver
> [el plan de implementación](docs/plan/plan-implementacion.md) para el
> resto de hitos.

## Instalar (usuario final)

1. Descarga `Arantxator-Setup.exe` (generado con `scripts/build.ps1`, ver
   más abajo) y haz doble clic.
2. Sigue el asistente — no requiere conexión a internet ni permisos de
   administrador. Deja marcada la casilla para crear el acceso directo en
   el escritorio.
3. Al abrirlo, la aplicación arranca en segundo plano y el navegador se
   abre solo en `http://127.0.0.1:8080`.
4. **Copia de seguridad:** cerrar la app y copiar `arantxator.db` (junto al
   ejecutable) es la copia completa — datos y documentos incluidos.

Detalle completo del proceso de instalación, desinstalación y solución de
problemas: [`docs/despliegue/instalacion-despliegue.md`](docs/despliegue/instalacion-despliegue.md).

## Qué es

Una única persona gestiona su cartera de inmuebles en alquiler (mono-cartera,
mono-usuario) desde una interfaz web local: altas de propiedades e
inquilinos, seguimiento de contratos (con las reglas de la LAU para la
Comunidad de Madrid), registro de gastos con reparto entre inquilinos en
pisos compartidos, y gestión de incidencias de mantenimiento.

El análisis de requisitos completo y el diseño técnico/funcional están en
[`docs/design/diseno-tecnico-funcional.md`](docs/design/diseno-tecnico-funcional.md).

## Stack

- **Backend:** Go — compila a un único binario nativo, sin runtime que instalar.
- **Base de datos:** SQLite embebida (driver puro Go, sin CGO). Los documentos
  adjuntos (contratos, facturas, fotos) se guardan como BLOB dentro de la
  propia base de datos, así que una copia de seguridad es siempre un único
  fichero `.db`.
- **GUI:** SPA web, compilada y embebida en el binario con `go:embed` — al
  arrancar el ejecutable, la interfaz, la API y los datos ya están operativos
  sin pasos de instalación adicionales.

## Estructura del repositorio

```
cmd/arantxator/          Punto de entrada del binario (+ icono/versión embebidos)
internal/
  config/                 Configuración de arranque (puerto, ruta de la BD)
  domain/                 Entidades: Inmueble, Habitacion, Inquilino, Contrato,
                          Gasto, RepartoGasto, Incidencia, Documento
  httpapi/                Rutas y handlers de la API HTTP interna
  storage/sqlite/          Conexión SQLite + migraciones embebidas
  webui/                    GUI embebida (SPA compilada, servida con fallback
                            de rutas para que React Router funcione al
                            recargar o abrir un enlace directo)
web/                      Código fuente de la SPA (React + Vite + TypeScript)
docs/design/               Análisis de requisitos y diseño técnico/funcional
docs/plan/                  Plan de implementación por hitos y su batería de pruebas
docs/despliegue/             Instalación y despliegue (build, instalador, icono)
installer/                    Script de Inno Setup e icono de la aplicación
scripts/                       Scripts de compilación (SPA + binario + instalador)
```

## Compilar y ejecutar (desarrollo)

Requiere [Go](https://go.dev/dl/) 1.23+ y [Node.js](https://nodejs.org/) LTS.

```bash
go mod tidy
go run ./cmd/arantxator
```

Arranca un servidor local y abre el navegador automáticamente en
`http://127.0.0.1:8080`, sirviendo la SPA ya compilada que va embebida en
`internal/webui/dist/`. Para trabajar en la SPA con recarga en caliente,
en otra terminal: `cd web && npm install && npm run dev` (proxy hacia la
API real en `:8080`).

Para generar el binario final y el instalador de Windows:

```powershell
powershell -File scripts/build.ps1
```

Detalle completo (icono, versión, Inno Setup, verificación realizada):
[`docs/despliegue/instalacion-despliegue.md`](docs/despliegue/instalacion-despliegue.md).

## Flujo de trabajo

- `development` es la rama base de trabajo; `main` solo recibe la versión
  inicial ya validada.
- Cada iteración se desarrolla en su propia rama y se integra en
  `development` mediante pull request.

## Próximos pasos

Módulo de Inquilinos (Hito 2 del [plan de implementación](docs/plan/plan-implementacion.md)) — ficha del inquilino,
documentación adjunta, y la asignación de ocupante a las habitaciones de
un inmueble compartido que el Hito 1 dejó preparada.
