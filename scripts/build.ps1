# Compila el backend Go y genera el ejecutable autocontenido de
# Arantxator Flat Admin.
#
# Requiere Go instalado (https://go.dev/dl/). Cuando exista la SPA en web/,
# este script tambien debera compilarla antes (npm run build) para que
# internal/webui/dist quede actualizado antes del go build.
#
# Uso: powershell -File scripts/build.ps1

$ErrorActionPreference = "Stop"

Write-Host "Compilando Arantxator Flat Admin..."
go build -o bin/arantxator.exe ./cmd/arantxator

Write-Host "Listo: bin/arantxator.exe"
