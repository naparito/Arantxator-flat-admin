# web/

SPA de Arantxator Flat Admin (React + Vite + TypeScript). Sin librería de
componentes: CSS plano con el sistema de tokens de los mockups aprobados
(`src/styles/tokens.css` + `src/styles/app.css`).

```bash
npm install
npm run dev     # servidor de Vite en :5173, con proxy /api -> http://127.0.0.1:8080
npm run build   # genera internal/webui/dist/, que go:embed incluye en el binario
npm test        # Vitest + React Testing Library
npm run lint    # oxlint (solo avisos; no forma parte de la batería de cierre)
```

Para trabajar con recarga en caliente contra la API real, arranca el backend en
otra terminal (`go run ./cmd/arantxator`) y luego `npm run dev` aquí: las
peticiones a `/api/*` se redirigen automáticamente a `:8080` (ver
`vite.config.ts`).

El resultado de `npm run build` se genera directamente en
`internal/webui/dist/`, la carpeta que el backend Go embebe en el binario final
con `go:embed` (ver `internal/webui/embed.go`). **Ese directorio se versiona en
el repositorio** para que `go build` funcione por sí solo sin depender de tener
Node instalado — hay que confirmar el `dist/` regenerado junto a los cambios de
`src/`.

## Estructura

```
src/
  main.tsx              Punto de entrada (BrowserRouter + App)
  App.tsx               Rutas de la SPA + NotificacionesProvider
  api/
    client.ts           Cliente HTTP tipado de /api/*
    types.ts            Tipos de todas las entidades y helpers de fecha
    notificacionesContext.tsx   Contador de avisos sin leer del rail
  components/
    Layout.tsx          Shell: barra lateral + contenido
    Sidebar.tsx         Rail de navegación (Resumen, Notificaciones, Cartera)
    AvisoArranque.tsx   Guard del dashboard: aviso-resumen al arrancar
    icons.tsx           Iconos SVG inline
    EstadoPill.tsx      Pill de estado reutilizable
  pages/                Una pantalla por fichero:
    Resumen · Notificaciones
    InmueblesListado · InmueblesFicha · InmuebleForm
    InquilinosListado · InquilinosFicha · InquilinoForm
    ContratosListado · ContratosFicha · ContratoForm
    Gastos
    *Util.tsx           Helpers y sub-componentes de cada pantalla
  styles/
    tokens.css          Paleta y tipografía (Instrument Sans + Public Sans)
    app.css             Clases de layout y componentes
  test/setup.ts         Configuración de Vitest (jest-dom)
```

Cada pantalla con lógica propia tiene su fichero `*.test.tsx` al lado.
Rutas: ver `App.tsx`. Endpoints que consume: ver `api/client.ts` y el
[manual funcional](../docs/manual/README.md#todos-los-endpoints-de-la-api).
