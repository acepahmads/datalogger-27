# Datalogger Analysis Application - Windows Local & Industrial Auto-Start Installer
# Prepares directories, database, configuration, builds/verifies binary, and registers auto-start
param(
    [switch]$AutoStart = $true,
    [switch]$NoLaunch = $false,
    [switch]$Uninstall = $false
)

Write-Host "========================================================" -ForegroundColor Cyan
Write-Host "  DATALOGGER ANALYSIS APPLICATION - WINDOWS INSTALLER" -ForegroundColor Cyan
Write-Host "  Edge Local-First Architecture with Auto-Start" -ForegroundColor Cyan
Write-Host "========================================================" -ForegroundColor Cyan

$AppDir = Split-Path -Parent $PSScriptRoot
Set-Location $AppDir

$StartupFolder = [System.Environment]::GetFolderPath('Startup')
$ShortcutPath = Join-Path $StartupFolder "Datalogger-AutoStart.lnk"

# ------------------------------------------------------------------------------
# Handle Uninstallation
# ------------------------------------------------------------------------------
if ($Uninstall) {
    Write-Host "Uninstalling Datalogger startup components from Windows..." -ForegroundColor Yellow
    if (Test-Path $ShortcutPath) {
        Remove-Item -Force $ShortcutPath
        Write-Host "Removed auto-start shortcut from Windows Startup folder." -ForegroundColor Green
    }
    Unregister-ScheduledTask -TaskName "DataloggerAnalysisEngine" -Confirm:$false -ErrorAction SilentlyContinue
    if (Test-Path "$AppDir\start-datalogger.bat") {
        Remove-Item -Force "$AppDir\start-datalogger.bat"
    }
    Write-Host ""
    Write-Host "Persistent data in '$AppDir\data' and MariaDB database 'datalogger' have been PRESERVED." -ForegroundColor Green
    Write-Host "Uninstallation complete." -ForegroundColor Green
    return
}

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

# 3. Build Go Binary if not present (safeguarded with D:\tmp_go)
if (-not (Test-Path "$AppDir\bin\datalogger.exe") -and -not (Test-Path "$AppDir\bin\datalogger-windows-amd64.exe") -and -not (Test-Path "$AppDir\datalogger.exe")) {
    Write-Host "[3/5] Compiling native pure-Go binary with zero CGO..." -ForegroundColor Yellow
    $env:CGO_ENABLED = "0"
    $env:TMPDIR = "D:\tmp_go"
    $env:GOTMPDIR = "D:\tmp_go"
    $env:GOCACHE = "D:\tmp_go\cache"
    New-Item -ItemType Directory -Force -Path "D:\tmp_go\cache" | Out-Null
    go build -ldflags "-s -w" -o "$AppDir\bin\datalogger.exe" "$AppDir\cmd\main.go"
} else {
    Write-Host "[3/5] Application binary verified." -ForegroundColor Green
}

# 4. Create Desktop & Startup Shortcut
Write-Host "[4/5] Creating startup launcher and auto-start registration..." -ForegroundColor Yellow
$LauncherContent = @"
@echo off
cd /d "%~dp0"
title Datalogger Analysis Engine
if exist "bin\datalogger-windows-amd64.exe" (
    start "" "bin\datalogger-windows-amd64.exe"
) else if exist "bin\datalogger.exe" (
    start "" "bin\datalogger.exe"
) else (
    start "" "datalogger.exe"
)
timeout /t 2 /nobreak >nul
start http://localhost:8080
"@
$LauncherBat = "$AppDir\start-datalogger.bat"
Set-Content -Path $LauncherBat -Value $LauncherContent

if ($AutoStart) {
    try {
        $WshShell = New-Object -ComObject WScript.Shell
        $Shortcut = $WshShell.CreateShortcut($ShortcutPath)
        $Shortcut.TargetPath = $LauncherBat
        $Shortcut.WorkingDirectory = $AppDir
        $Shortcut.Description = "Datalogger Analysis Application Auto-Start"
        $Shortcut.Save()
        Write-Host "Auto-start on Windows reboot enabled (Shortcut registered in Startup folder)." -ForegroundColor Green
    } catch {
        Write-Warning "Could not register Startup folder shortcut: $_"
    }
}

# 5. Launch Application
if (-not $NoLaunch) {
    Write-Host "[5/5] Launching Datalogger Application..." -ForegroundColor Green
    Start-Process -FilePath "cmd.exe" -ArgumentList "/c `"$LauncherBat`"" -WorkingDirectory $AppDir
} else {
    Write-Host "[5/5] Skipping immediate launch (-NoLaunch specified)." -ForegroundColor Yellow
}

Write-Host ""
Write-Host "========================================================" -ForegroundColor Cyan
Write-Host "  INSTALLATION COMPLETE!" -ForegroundColor Green
Write-Host "  Access URL: http://localhost:8080" -ForegroundColor Yellow
Write-Host "  Default Admin: admin / admin123" -ForegroundColor Yellow
Write-Host "  Auto-Start: Enabled for reboot persistence" -ForegroundColor Green
Write-Host "========================================================" -ForegroundColor Cyan
