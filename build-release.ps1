$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location -LiteralPath $root

$oldGOOS = $env:GOOS
$oldGOARCH = $env:GOARCH
$oldCGO_ENABLED = $env:CGO_ENABLED
$oldGOCACHE = $env:GOCACHE

$cacheDir = Join-Path $root ".gocache"
$toolsDir = Join-Path $root ".tools\bin"
$versionInfoTool = Join-Path $toolsDir "goversioninfo.exe"
$iconPath = Join-Path $root "winproxy.ico"
$versionInfoPath = Join-Path $root "versioninfo.json"
$sysoPath = Join-Path $root "resource_windows_386.syso"

$appVersion = "1.0.0.0"
$appName = "WinProxy"
$appDescription = "Remote Desktop SOCKS5 Proxy"
$appCompany = "WinProxy"
$appCopyright = "Copyright (c) 2026 WinProxy"

New-Item -ItemType Directory -Force -Path $cacheDir | Out-Null
New-Item -ItemType Directory -Force -Path $toolsDir | Out-Null

function Invoke-Go {
    param(
        [Parameter(Mandatory = $true)]
        [string[]]$Args
    )

    & go @Args
    if ($LASTEXITCODE -ne 0) {
        throw "go $($Args -join ' ') failed with exit code $LASTEXITCODE."
    }
}

function Ensure-AppIcon {
    param([string]$OutputPath)

    if (Test-Path -LiteralPath $OutputPath) {
        return
    }

    Add-Type -AssemblyName System.Drawing
    $sizes = @(16, 32, 48, 256)
    $images = @()

    foreach ($size in $sizes) {
        $bmp = [System.Drawing.Bitmap]::new($size, $size, [System.Drawing.Imaging.PixelFormat]::Format32bppArgb)
        $g = [System.Drawing.Graphics]::FromImage($bmp)
        $g.SmoothingMode = [System.Drawing.Drawing2D.SmoothingMode]::AntiAlias

        $rect = [System.Drawing.Rectangle]::new(0, 0, $size - 1, $size - 1)
        $brush = [System.Drawing.Drawing2D.LinearGradientBrush]::new(
            $rect,
            [System.Drawing.Color]::FromArgb(18, 101, 242),
            [System.Drawing.Color]::FromArgb(42, 164, 255),
            45
        )

        $radius = [float]([Math]::Max(4, $size / 4))
        $shapePath = [System.Drawing.Drawing2D.GraphicsPath]::new()
        $diameter = $radius * 2
        $shapePath.AddArc(0, 0, $diameter, $diameter, 180, 90)
        $shapePath.AddArc($size - $diameter - 1, 0, $diameter, $diameter, 270, 90)
        $shapePath.AddArc($size - $diameter - 1, $size - $diameter - 1, $diameter, $diameter, 0, 90)
        $shapePath.AddArc(0, $size - $diameter - 1, $diameter, $diameter, 90, 90)
        $shapePath.CloseFigure()
        $g.FillPath($brush, $shapePath)

        $font = [System.Drawing.Font]::new("Segoe UI", [single]($size * 0.58), [System.Drawing.FontStyle]::Bold, [System.Drawing.GraphicsUnit]::Pixel)
        $sf = [System.Drawing.StringFormat]::new()
        $sf.Alignment = [System.Drawing.StringAlignment]::Center
        $sf.LineAlignment = [System.Drawing.StringAlignment]::Center
        $g.DrawString("W", $font, [System.Drawing.Brushes]::White, [System.Drawing.RectangleF]::new(0, -1, $size, $size), $sf)

        $png = [System.IO.MemoryStream]::new()
        $bmp.Save($png, [System.Drawing.Imaging.ImageFormat]::Png)
        $images += [pscustomobject]@{ Size = $size; Bytes = $png.ToArray() }

        $sf.Dispose()
        $font.Dispose()
        $shapePath.Dispose()
        $brush.Dispose()
        $g.Dispose()
        $bmp.Dispose()
    }

    $fs = [System.IO.File]::Create($OutputPath)
    $bw = [System.IO.BinaryWriter]::new($fs)
    $bw.Write([uint16]0)
    $bw.Write([uint16]1)
    $bw.Write([uint16]$images.Count)
    $offset = 6 + ($images.Count * 16)

    foreach ($img in $images) {
        $dim = $img.Size
        if ($dim -eq 256) {
            $dim = 0
        }
        $bw.Write([byte]$dim)
        $bw.Write([byte]$dim)
        $bw.Write([byte]0)
        $bw.Write([byte]0)
        $bw.Write([uint16]1)
        $bw.Write([uint16]32)
        $bw.Write([uint32]$img.Bytes.Length)
        $bw.Write([uint32]$offset)
        $offset += $img.Bytes.Length
    }

    foreach ($img in $images) {
        $bw.Write($img.Bytes)
    }

    $bw.Dispose()
    $fs.Dispose()
}

function Ensure-VersionInfoTool {
    if (Test-Path -LiteralPath $versionInfoTool) {
        return
    }

    Write-Host "Installing goversioninfo..."
    $env:GOBIN = $toolsDir
    Invoke-Go @("install", "github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest")
}

function Write-VersionInfoJson {
    $payload = [ordered]@{
        FixedFileInfo = [ordered]@{
            FileVersion    = [ordered]@{ Major = 1; Minor = 0; Patch = 0; Build = 0 }
            ProductVersion = [ordered]@{ Major = 1; Minor = 0; Patch = 0; Build = 0 }
            FileFlagsMask  = "3f"
            FileFlags      = "00"
            FileOS         = "040004"
            FileType       = "01"
            FileSubType    = "00"
        }
        StringFileInfo = [ordered]@{
            Comments         = "Local Remote Desktop forwarding through SOCKS5."
            CompanyName      = $appCompany
            FileDescription  = $appDescription
            FileVersion      = $appVersion
            InternalName     = "winproxy"
            LegalCopyright   = $appCopyright
            OriginalFilename = "winproxy.exe"
            ProductName      = $appName
            ProductVersion   = $appVersion
        }
        VarFileInfo = [ordered]@{
            Translation = [ordered]@{
                LangID    = "0409"
                CharsetID = "04B0"
            }
        }
        IconPath = $iconPath
        ManifestPath = ""
    }

    $json = $payload | ConvertTo-Json -Depth 8
    [System.IO.File]::WriteAllText($versionInfoPath, $json, [System.Text.UTF8Encoding]::new($false))
}

function Build-Resources {
    Ensure-AppIcon -OutputPath $iconPath
    Ensure-VersionInfoTool
    Write-VersionInfoJson

    Get-ChildItem -LiteralPath $root -Filter "resource_windows_*.syso" -ErrorAction SilentlyContinue | Remove-Item -Force

    & $versionInfoTool -platform-specific=true -o $sysoPath $versionInfoPath
    if ($LASTEXITCODE -ne 0) {
        throw "goversioninfo failed with exit code $LASTEXITCODE."
    }
    Get-ChildItem -LiteralPath $root -Filter "resource_windows_*.syso" -ErrorAction SilentlyContinue |
        Where-Object { $_.Name -ne "resource_windows_386.syso" } |
        Remove-Item -Force
}

function Build-WinProxy {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Arch,

        [Parameter(Mandatory = $true)]
        [string]$Output
    )

    Write-Host "Building $Output ($Arch)..."
    $env:GOOS = "windows"
    $env:GOARCH = $Arch
    $env:CGO_ENABLED = "0"
    $env:GOCACHE = $cacheDir

    Build-Resources
    Invoke-Go @("build", "-buildvcs=false", "-ldflags=-H windowsgui", "-o", $Output, ".")

    $file = Get-Item -LiteralPath $Output
    Write-Host ("OK: {0}  {1:N0} bytes" -f $file.Name, $file.Length)
}

try {
    if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
        throw "Go is not installed or go.exe is not in PATH."
    }

    $env:GOCACHE = $cacheDir
    Invoke-Go @("test", "./...")

    Build-WinProxy -Arch "386" -Output "winproxy.exe"

    Write-Host ""
    Write-Host "Build complete."
} finally {
    $env:GOOS = $oldGOOS
    $env:GOARCH = $oldGOARCH
    $env:CGO_ENABLED = $oldCGO_ENABLED
    $env:GOCACHE = $oldGOCACHE
}
