# Compila el backend Go y genera el ejecutable autocontenido de
# Arantxator Flat Admin.
#
# Requiere Go y Node.js instalados. Primero compila la SPA (web/) con
# npm run build, que deja el resultado en internal/webui/dist/ para que
# go:embed lo incluya en el binario final.
#
# Uso: powershell -File scripts/build.ps1

$ErrorActionPreference = "Stop"

Write-Host "Compilando la SPA..."
Push-Location web
npm install
npm run build
Pop-Location

Write-Host "Compilando Arantxator Flat Admin..."
go build -o bin/arantxator.exe ./cmd/arantxator

Write-Host "Listo: bin/arantxator.exe"
