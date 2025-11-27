# Створення іконки для LockScreen Sync
# 32x32 ICO файл з зображенням монітора

$iconPath = Join-Path $PSScriptRoot "icon.ico"

# ICO header
$ico = [System.Collections.Generic.List[byte]]::new()

# ICO Header (6 bytes)
$ico.AddRange([byte[]](0x00, 0x00))  # Reserved
$ico.AddRange([byte[]](0x01, 0x00))  # Type (1 = ICO)
$ico.AddRange([byte[]](0x01, 0x00))  # Number of images

# ICO Directory Entry (16 bytes)
$width = 32
$height = 32
$bmpDataSize = 40 + ($width * $height * 4) + ($width * $height / 8)

$ico.Add([byte]$width)               # Width
$ico.Add([byte]$height)              # Height
$ico.Add([byte]0x00)                 # Color palette
$ico.Add([byte]0x00)                 # Reserved
$ico.AddRange([byte[]](0x01, 0x00))  # Color planes
$ico.AddRange([byte[]](0x20, 0x00))  # Bits per pixel (32)
$ico.AddRange([System.BitConverter]::GetBytes([int]$bmpDataSize))  # Size
$ico.AddRange([System.BitConverter]::GetBytes([int]22))  # Offset

# BITMAPINFOHEADER (40 bytes)
$ico.AddRange([System.BitConverter]::GetBytes([int]40))  # Header size
$ico.AddRange([System.BitConverter]::GetBytes([int]$width))  # Width
$ico.AddRange([System.BitConverter]::GetBytes([int]($height * 2)))  # Height (doubled)
$ico.AddRange([byte[]](0x01, 0x00))  # Planes
$ico.AddRange([byte[]](0x20, 0x00))  # Bits per pixel
$ico.AddRange([System.BitConverter]::GetBytes([int]0))  # Compression
$ico.AddRange([System.BitConverter]::GetBytes([int]0))  # Image size
$ico.AddRange([System.BitConverter]::GetBytes([int]0))  # X ppm
$ico.AddRange([System.BitConverter]::GetBytes([int]0))  # Y ppm
$ico.AddRange([System.BitConverter]::GetBytes([int]0))  # Colors used
$ico.AddRange([System.BitConverter]::GetBytes([int]0))  # Important colors

# Pixel data (BGRA, bottom-up)
for ($y = $height - 1; $y -ge 0; $y--) {
    for ($x = 0; $x -lt $width; $x++) {
        # Монітор
        $isFrame = ($x -ge 4 -and $x -le 27 -and $y -ge 6 -and $y -le 21) -and -not ($x -ge 6 -and $x -le 25 -and $y -ge 8 -and $y -le 19)
        $isScreen = $x -ge 6 -and $x -le 25 -and $y -ge 8 -and $y -le 19
        $isStand = $x -ge 13 -and $x -le 18 -and $y -ge 22 -and $y -le 25
        $isBase = $x -ge 9 -and $x -le 22 -and $y -ge 26 -and $y -le 27

        # Градієнт на екрані (імітація шпалер)
        $screenGradient = [math]::Floor(($x - 6) / 20.0 * 100 + ($y - 8) / 12.0 * 50)

        if ($isScreen) {
            # Градієнт синьо-фіолетовий (як шпалери Windows)
            $b = [math]::Min(255, 180 + $screenGradient / 3)
            $g = [math]::Min(255, 100 + $screenGradient / 4)
            $r = [math]::Min(255, 80 + $screenGradient / 2)
            $a = 255
        } elseif ($isFrame -or $isStand -or $isBase) {
            # Темно-сірий корпус
            $b = 60; $g = 60; $r = 60; $a = 255
        } else {
            # Прозорий фон
            $b = 0; $g = 0; $r = 0; $a = 0
        }

        $ico.AddRange([byte[]]($b, $g, $r, $a))
    }
}

# AND mask
for ($i = 0; $i -lt ($width * $height / 8); $i++) {
    $ico.Add([byte]0x00)
}

[System.IO.File]::WriteAllBytes($iconPath, $ico.ToArray())
Write-Host "Icon created: $iconPath" -ForegroundColor Green
