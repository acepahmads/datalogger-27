# ==============================================================================
# Datalogger Analysis Application - Multi-Platform Distribution Packager
# Packages release archives (.tar.gz and .zip) with binaries, Web UI, and service scripts
# ==============================================================================

param(
    [string]$Version = "v1.0.0-phase9",
    [string]$OutputDir = "release"
)

$ErrorActionPreference = "Stop"
$AppDir = Split-Path -Parent $PSScriptRoot
Set-Location $AppDir

Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "  DATALOGGER MULTI-PLATFORM RELEASE PACKAGER ($Version)" -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan

$ReleasePath = Join-Path $AppDir $OutputDir
if (Test-Path $ReleasePath) {
    Remove-Item -Recurse -Force $ReleasePath
}
New-Item -ItemType Directory -Force -Path $ReleasePath | Out-Null

$DistWeb = Join-Path $AppDir "web\dist"
if (-not (Test-Path (Join-Path $DistWeb "index.html"))) {
    Write-Host "Building Web UI distribution bundle..." -ForegroundColor Yellow
    npm --prefix "$AppDir\web" run build
}

# Platform targets definition
$Targets = @(
    @{
        Name = "windows-amd64"
        Binary = "bin\datalogger-windows-amd64.exe"
        ArchiveType = "zip"
        ExecutableName = "datalogger.exe"
        Installer = "scripts\install-windows.ps1"
    },
    @{
        Name = "linux-amd64"
        Binary = "bin\datalogger-linux-amd64"
        ArchiveType = "tar.gz"
        ExecutableName = "datalogger"
        Installer = "scripts\install-linux.sh"
    },
    @{
        Name = "linux-arm64"
        Binary = "bin\datalogger-linux-arm64"
        ArchiveType = "tar.gz"
        ExecutableName = "datalogger"
        Installer = "scripts\install-linux.sh"
    },
    @{
        Name = "linux-armv7"
        Binary = "bin\datalogger-linux-armv7"
        ArchiveType = "tar.gz"
        ExecutableName = "datalogger"
        Installer = "scripts\install-linux.sh"
    },
    @{
        Name = "linux-armv6"
        Binary = "bin\datalogger-linux-armv6"
        ArchiveType = "tar.gz"
        ExecutableName = "datalogger"
        Installer = "scripts\install-linux.sh"
    },
    @{
        Name = "darwin-amd64"
        Binary = "bin\datalogger-darwin-amd64"
        ArchiveType = "tar.gz"
        ExecutableName = "datalogger"
        Installer = "scripts\install-darwin.sh"
    },
    @{
        Name = "darwin-arm64"
        Binary = "bin\datalogger-darwin-arm64"
        ArchiveType = "tar.gz"
        ExecutableName = "datalogger"
        Installer = "scripts\install-darwin.sh"
    }
)

$ChecksumLines = @()

foreach ($t in $Targets) {
    $PlatformName = $t.Name
    $BinPath = Join-Path $AppDir $t.Binary
    if (-not (Test-Path $BinPath)) {
        Write-Warning "Binary not found: $BinPath. Skipping $PlatformName."
        continue
    }

    Write-Host "Packaging platform: $PlatformName..." -ForegroundColor Yellow
    $StageDir = Join-Path $AppDir "release_stage_$PlatformName"
    if (Test-Path $StageDir) { Remove-Item -Recurse -Force $StageDir }
    New-Item -ItemType Directory -Force -Path $StageDir | Out-Null
    New-Item -ItemType Directory -Force -Path (Join-Path $StageDir "web") | Out-Null

    # Copy binary
    Copy-Item -Path $BinPath -Destination (Join-Path $StageDir $t.ExecutableName)
    
    # Copy web assets
    Copy-Item -Recurse -Path $DistWeb -Destination (Join-Path $StageDir "web\dist")

    # Copy config.json
    if (Test-Path (Join-Path $AppDir "config.json")) {
        Copy-Item -Path (Join-Path $AppDir "config.json") -Destination (Join-Path $StageDir "config.json")
    }

    # Copy installers and helpers
    if ($t.Installer -and (Test-Path (Join-Path $AppDir $t.Installer))) {
        Copy-Item -Path (Join-Path $AppDir $t.Installer) -Destination $StageDir
    }
    if ($PlatformName -like "linux*") {
        if (Test-Path (Join-Path $AppDir "scripts\update-linux.sh")) {
            Copy-Item -Path (Join-Path $AppDir "scripts\update-linux.sh") -Destination $StageDir
        }
        if (Test-Path (Join-Path $AppDir "scripts\setup-usb-udev.sh")) {
            Copy-Item -Path (Join-Path $AppDir "scripts\setup-usb-udev.sh") -Destination $StageDir
        }
    } elseif ($PlatformName -like "darwin*") {
        if (Test-Path (Join-Path $AppDir "scripts\com.datalogger.analysis.plist")) {
            Copy-Item -Path (Join-Path $AppDir "scripts\com.datalogger.analysis.plist") -Destination $StageDir
        }
    }

    # Archive creation
    $ArchiveBase = "datalogger-$Version-$PlatformName"
    if ($t.ArchiveType -eq "zip") {
        $ArchiveFile = Join-Path $ReleasePath "$ArchiveBase.zip"
        Compress-Archive -Path "$StageDir\*" -DestinationPath $ArchiveFile -Force
        Write-Host "  -> Created: $ArchiveFile" -ForegroundColor Green
    } else {
        $ArchiveFile = Join-Path $ReleasePath "$ArchiveBase.tar.gz"
        # Use tar utility available in Windows 10/11 & Linux
        tar -czf "$ArchiveFile" -C "$StageDir" .
        Write-Host "  -> Created: $ArchiveFile" -ForegroundColor Green
    }

    # Clean up stage dir
    Remove-Item -Recurse -Force $StageDir

    # Compute SHA256
    $Hash = (Get-FileHash -Path $ArchiveFile -Algorithm SHA256).Hash.ToLower()
    $Filename = Split-Path -Leaf $ArchiveFile
    $ChecksumLines += "$Hash  $Filename"
}

# Write SHA256SUMS.txt
$ChecksumPath = Join-Path $ReleasePath "SHA256SUMS.txt"
$ChecksumLines | Out-File -FilePath $ChecksumPath -Encoding utf8
Write-Host ""
Write-Host "SHA256 Checksums written to: $ChecksumPath" -ForegroundColor Cyan
Get-Content $ChecksumPath

Write-Host ""
Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "  ALL PACKAGES CREATED SUCCESSFULLY IN: $ReleasePath" -ForegroundColor Green
Write-Host "==========================================================" -ForegroundColor Cyan
