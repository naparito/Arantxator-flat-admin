# Arantxator Flat Admin

Sistema de gestión integral de alquileres: inmuebles, inquilinos, contratos,
gastos con reparto porcentual e incidencias, en una aplicación visual,
autocontenida y pensada para un usuario sin conocimientos técnicos.

> **Estado:** scaffold inicial. Sin funcionalidad todavía — ver
> [Próximos pasos](#próximos-pasos).

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
cmd/arantxator/         Punto de entrada del binario
internal/
  config/                Configuración de arranque (puerto, ruta de la BD)
  domain/                Entidades: Inmueble, Inquilino, Contrato, Gasto,
                          RepartoGasto, Incidencia, Documento
  httpapi/                Rutas y handlers de la API HTTP interna
  storage/sqlite/          Conexión SQLite + migraciones embebidas
  webui/                    GUI embebida (dist/ se sustituirá por la SPA)
web/                     Código fuente de la futura SPA (pendiente de mockups)
docs/design/              Análisis de requisitos y diseño técnico/funcional
installer/                 Empaquetado en instalador de Windows (pendiente)
scripts/                    Scripts de compilación
```

## Compilar y ejecutar

Requiere [Go](https://go.dev/dl/) 1.23 o superior.

```bash
go mod tidy
go run ./cmd/arantxator
```

Arranca un servidor local y abre el navegador automáticamente en
`http://127.0.0.1:8080`.

Para generar el ejecutable:

```powershell
powershell -File scripts/build.ps1
```

## Flujo de trabajo

- `development` es la rama base de trabajo; `main` solo recibe la versión
  inicial ya validada.
- Cada iteración se desarrolla en su propia rama y se integra en
  `development` mediante pull request.

## Próximos pasos

1. `go mod tidy` para fijar la dependencia de SQLite (`modernc.org/sqlite`).
2. Mockups de la GUI (limpia, colorida, minimalista) antes de implementar
   pantallas.
3. Módulo de Inmuebles end-to-end (API + GUI) como primera vertical completa.
