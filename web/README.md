# web/

Código fuente de la SPA (a definir en el pase de mockups — GUI limpia,
colorida y minimalista, pendiente de arrancar).

El resultado de `npm run build` debe generarse directamente en
`internal/webui/dist/`, que es la carpeta que el backend Go embebe en el
binario final con `go:embed` (ver `internal/webui/embed.go`). Hasta entonces,
`internal/webui/dist/index.html` sirve como página mínima para comprobar que
el servidor funciona.
