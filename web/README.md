# web/

SPA de Arantxator Flat Admin (React + Vite + TypeScript), bootstrap del
Hito 1. Sin librería de componentes: CSS plano con el sistema de tokens de
los mockups aprobados (`src/styles/tokens.css`).

```bash
npm install
npm run dev     # servidor de Vite en :5173, con proxy /api -> http://127.0.0.1:8080
npm run build   # genera internal/webui/dist/, que go:embed incluye en el binario
npm test        # Vitest + React Testing Library
```

Para trabajar con recarga en caliente contra la API real, arranca el
backend en otra terminal (`go run ./cmd/arantxator`) y luego `npm run dev`
aquí: las peticiones a `/api/*` se redirigen automáticamente a `:8080`
(ver `vite.config.ts`).

El resultado de `npm run build` se genera directamente en
`internal/webui/dist/`, la carpeta que el backend Go embebe en el binario
final con `go:embed` (ver `internal/webui/embed.go`). Ese directorio se
versiona en el repositorio para que `go build` funcione por sí solo sin
depender de tener Node instalado.
