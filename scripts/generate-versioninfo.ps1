# Genera cmd/arantxator/resource_windows_*.syso a partir de
# cmd/arantxator/versioninfo.json e installer/icon.ico, para que el .exe
# final lleve incrustados el icono de la aplicacion y los metadatos de
# version (nombre, version, copyright) que se ven en las Propiedades del
# fichero en el Explorador de Windows.
#
# go build recoge automaticamente los .syso que encuentre en el paquete que
# compila (comportamiento nativo del compilador en windows/*, sin flags
# especiales) — por eso los .syso generados se versionan en el repositorio:
# un "go build" normal en una maquina limpia ya produce un .exe con icono,
# sin depender de tener goversioninfo instalado.
#
# Solo hace falta volver a ejecutar este script si cambia el icono
# (installer/icon.ico) o los metadatos de versioninfo.json.
#
# Uso: powershell -File scripts/generate-versioninfo.ps1

$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
$goverPath = Join-Path (& go env GOPATH) "bin\goversioninfo.exe"

if (-not (Test-Path $goverPath)) {
    Write-Host "goversioninfo no esta instalado; instalando..."
    go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest
}

Push-Location (Join-Path $repoRoot "cmd\arantxator")
try {
    & $goverPath -platform-specific=true -o resource.syso versioninfo.json
    Write-Host "Generados los resource_windows_*.syso en cmd/arantxator/"
} finally {
    Pop-Location
}
