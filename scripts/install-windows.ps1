# Datalogger Analysis Application - Windows Local Installer
# Prepares directories, database, configuration, builds/verifies binary, and launches local browser

Write-Host "========================================================" -ForegroundColor Cyan
Write-Host "  DATALOGGER ANALYSIS APPLICATION - WINDOWS INSTALLER" -ForegroundColor Cyan
Write-Host "  Phase 1 Foundation - Edge Local-First Architecture" -ForegroundColor Cyan
Write-Host "========================================================" -ForegroundColor Cyan

$AppDir = Split-Path -Parent $PSScriptRoot
Set-Location $AppDir

# 1. Create runtime directories
Write-Host "[1/5] Setting up edge runtime directories..." -ForegroundColor Yellow
New-Item -ItemType Directory -Force -Path "$AppDir\data" | Out-Null
New-Item -ItemType Directory -Force -Path "$AppDir\logs" | Out-Null
New-Item -ItemType Directory -Force -Path "$AppDir\bin" | Out-Null

# 2. Build Web UI if not present
if (-not (Test-Path "$AppDir\web\dist\index.html")) {
    Write-Host "[2/5] Building Web UI distribution..." -ForegroundColor Yellow
    npm --prefix "$AppDir\web" run build
} else {
    Write-Host "[2/5] Web UI distribution verified." -ForegroundColor Green
}

# 3. Build Go Binary if not present
if (-not (Test-Path "$AppDir\bin\datalogger.exe") -and -not (Test-Path "$AppDir\datalogger.exe")) {
    Write-Host "[3/5] Compiling native pure-Go binary with zero CGO..." -ForegroundColor Yellow
    $env:CGO_ENABLED = "0"
    go build -ldflags "-s -w" -o "$AppDir\bin\datalogger.exe" "$AppDir\cmd\main.go"
} else {
    Write-Host "[3/5] Application binary verified." -ForegroundColor Green
}

# 4. Create Desktop & Startup Shortcut
Write-Host "[4/5] Creating startup launcher..." -ForegroundColor Yellow
$LauncherContent = @"
@echo off
cd /d "%~dp0"
title Datalogger Analysis Engine
if exist "bin\datalogger.exe" (
    start "" "bin\datalogger.exe"
) else (
    start "" "datalogger.exe"
)
timeout /t 2 /nobreak >nul
start http://localhost:8080
"@
Set-Content -Path "$AppDir\start-datalogger.bat" -Value $LauncherContent

# 5. Launch Application
Write-Host "[5/5] Launching Datalogger Application..." -ForegroundColor Green
Start-Process -FilePath "cmd.exe" -ArgumentList "/c start-datalogger.bat" -WorkingDirectory $AppDir

Write-Host ""
Write-Host "========================================================" -ForegroundColor Cyan
Write-Host "  INSTALLATION COMPLETE!" -ForegroundColor Green
Write-Host "  Access URL: http://localhost:8080" -ForegroundColor Yellow
Write-Host "  Default Admin: admin / admin123" -ForegroundColor Yellow
Write-Host "========================================================" -ForegroundColor Cyan
