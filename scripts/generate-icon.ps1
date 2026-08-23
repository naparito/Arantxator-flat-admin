# Genera installer/icon.ico a partir del isotipo de Arantxator (el mismo
# glifo de casa usado en el rail de navegación de la SPA, ver
# web/src/components/Sidebar.tsx) dibujandolo directamente con
# System.Drawing — sin depender de un conversor SVG->ICO externo.
#
# Uso: powershell -File scripts/generate-icon.ps1
# Vuelve a ejecutarlo si cambia el isotipo de la marca.

$ErrorActionPreference = "Stop"
Add-Type -AssemblyName System.Drawing

$repoRoot = Split-Path -Parent $PSScriptRoot
$outPath = Join-Path $repoRoot "installer\icon.ico"

# Colores de marca (ver web/src/styles/tokens.css: --sidebar-bg, --sidebar-ink)
$bgColor = [System.Drawing.Color]::FromArgb(255, 0x23, 0x20, 0x1b)
$glyphColor = [System.Drawing.Color]::FromArgb(255, 0xF3, 0xEF, 0xE6)

# Glifo de casa tal como está en Sidebar.tsx, en un viewBox de 28x28:
#   tejado:   (4,13.5) -> (14,5) -> (24,13.5)
#   paredes:  (7,12) -> (7,22) -> (21,22) -> (21,12)
#   puerta:   (12,22) -> (12,16) -> (16,16) -> (16,22)
# stroke-width original = 2 en ese espacio de 28 unidades.
function New-IconBitmap([int]$size) {
    $bmp = New-Object System.Drawing.Bitmap $size, $size
    $g = [System.Drawing.Graphics]::FromImage($bmp)
    $g.SmoothingMode = [System.Drawing.Drawing2D.SmoothingMode]::AntiAlias
    $g.Clear([System.Drawing.Color]::Transparent)

    $pad = $size * 0.10
    $radius = $size * 0.22
    $rect = New-Object System.Drawing.RectangleF $pad, $pad, ($size - 2 * $pad), ($size - 2 * $pad)

    # Rectangulo redondeado de fondo (estilo icono moderno de Windows).
    $path = New-Object System.Drawing.Drawing2D.GraphicsPath
    $d = $radius * 2
    $path.AddArc($rect.X, $rect.Y, $d, $d, 180, 90)
    $path.AddArc($rect.Right - $d, $rect.Y, $d, $d, 270, 90)
    $path.AddArc($rect.Right - $d, $rect.Bottom - $d, $d, $d, 0, 90)
    $path.AddArc($rect.X, $rect.Bottom - $d, $d, $d, 90, 90)
    $path.CloseFigure()
    $brush = New-Object System.Drawing.SolidBrush $bgColor
    $g.FillPath($brush, $path)

    # Glifo de casa, escalado desde el viewBox de 28x28 al hueco interior.
    $inner = $size * 0.62
    $scale = $inner / 28.0
    $offset = ($size - 28 * $scale) / 2.0
    $strokeW = [Math]::Max(1.2, 2.0 * $scale)
    $pen = New-Object System.Drawing.Pen $glyphColor, $strokeW
    $pen.StartCap = [System.Drawing.Drawing2D.LineCap]::Round
    $pen.EndCap = [System.Drawing.Drawing2D.LineCap]::Round
    $pen.LineJoin = [System.Drawing.Drawing2D.LineJoin]::Round

    function P([double]$x, [double]$y) {
        New-Object System.Drawing.PointF (($x * $scale + $offset), ($y * $scale + $offset))
    }

    $roof = @(P 4 13.5; P 14 5; P 24 13.5)
    $walls = @(P 7 12; P 7 22; P 21 22; P 21 12)
    $door = @(P 12 22; P 12 16; P 16 16; P 16 22)

    $g.DrawLines($pen, $roof)
    $g.DrawLines($pen, $walls)
    $g.DrawLines($pen, $door)

    $pen.Dispose(); $brush.Dispose(); $path.Dispose(); $g.Dispose()
    return $bmp
}

$sizes = @(16, 32, 48, 256)
$pngBytesBySize = [ordered]@{}
foreach ($s in $sizes) {
    $bmp = New-IconBitmap $s
    $ms = New-Object System.IO.MemoryStream
    $bmp.Save($ms, [System.Drawing.Imaging.ImageFormat]::Png)
    $pngBytesBySize["$s"] = $ms.ToArray()
    $ms.Dispose(); $bmp.Dispose()
}

# Empaqueta las PNG en un .ico multi-resolución (formato soportado desde
# Windows Vista: cada entrada puede llevar PNG en vez de BMP crudo).
$fs = [System.IO.File]::Create($outPath)
$bw = New-Object System.IO.BinaryWriter $fs

$count = $sizes.Count
$bw.Write([UInt16]0)      # reserved
$bw.Write([UInt16]1)      # type = icon
$bw.Write([UInt16]$count)

$headerSize = 6 + 16 * $count
$offset = $headerSize
foreach ($s in $sizes) {
    $bytes = $pngBytesBySize["$s"]
    $dim = if ($s -ge 256) { 0 } else { $s }  # 0 significa 256 en el formato ICO
    $bw.Write([Byte]$dim)          # width
    $bw.Write([Byte]$dim)          # height
    $bw.Write([Byte]0)             # color palette
    $bw.Write([Byte]0)             # reserved
    $bw.Write([UInt16]1)           # color planes
    $bw.Write([UInt16]32)          # bits per pixel
    $bw.Write([UInt32]$bytes.Length)
    $bw.Write([UInt32]$offset)
    $offset += $bytes.Length
}
foreach ($s in $sizes) {
    $bw.Write($pngBytesBySize["$s"])
}

$bw.Flush(); $bw.Dispose(); $fs.Dispose()

Write-Host "Icono generado: $outPath ($($sizes -join ', ') px)"
