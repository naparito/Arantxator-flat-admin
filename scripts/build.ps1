# Compila el backend Go y genera el ejecutable autocontenido de
# Arantxator Flat Admin, y si Inno Setup esta instalado, tambien el
# instalador de Windows (dist/Arantxator-Setup.exe).
#
# Requiere Go y Node.js instalados. Primero compila la SPA (web/) con
# npm run build, que deja el resultado en internal/webui/dist/ para que
# go:embed lo incluya en el binario final. El icono y los metadatos de
# version del .exe ya vienen versionados como cmd/arantxator/resource_*.syso
# (ver scripts/generate-versioninfo.ps1 si hace falta regenerarlos).
#
# Uso: powershell -File scripts/build.ps1
#
# Ver docs/despliegue/instalacion-despliegue.md para el detalle completo
# del proceso de instalacion y despliegue.

$ErrorActionPreference = "Stop"
$repoRoot = Split-Path -Parent $PSScriptRoot

Write-Host "Compilando la SPA..."
Push-Location (Join-Path $repoRoot "web")
npm install
npm run build
Pop-Location

Write-Host "Compilando Arantxator Flat Admin..."
New-Item -ItemType Directory -Force -Path (Join-Path $repoRoot "bin") | Out-Null
go build -o (Join-Path $repoRoot "bin\arantxator.exe") (Join-Path $repoRoot "cmd\arantxator")
Write-Host "Listo: bin/arantxator.exe"

$isccPath = $null
$isccCmd = Get-Command ISCC -ErrorAction SilentlyContinue
if ($isccCmd) {
    $isccPath = $isccCmd.Source
} else {
    $candidate = Get-ChildItem "$env:LOCALAPPDATA\Programs\Inno Setup 6\ISCC.exe", "${env:ProgramFiles(x86)}\Inno Setup 6\ISCC.exe" -ErrorAction SilentlyContinue | Select-Object -First 1
    if ($candidate) { $isccPath = $candidate.FullName }
}

if ($isccPath) {
    Write-Host "Compilando el instalador..."
    & $isccPath (Join-Path $repoRoot "installer\setup.iss")
    Write-Host "Listo: dist/Arantxator-Setup.exe"
} else {
    Write-Host "Inno Setup no esta instalado, se omite el instalador."
    Write-Host "Instalalo con 'winget install JRSoftware.InnoSetup' y vuelve a ejecutar este script para generar dist/Arantxator-Setup.exe."
}
