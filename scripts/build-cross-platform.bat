@echo off
REM ========================================================
REM  Datalogger Cross-Platform Pure-Go Multi-Target Builder
REM  Requires NO CGO compiler (zero CGO pure Go SQLite & MariaDB)
REM ========================================================

set "TMPDIR=D:\tmp_go"
set "GOTMPDIR=D:\tmp_go"
set "GOCACHE=D:\tmp_go\cache"
if not exist "%TMPDIR%" mkdir "%TMPDIR%"
if not exist "%GOCACHE%" mkdir "%GOCACHE%"

echo Compiling cross-platform binaries into bin/...
if not exist "bin" mkdir "bin"

echo [1/7] Building Windows x64...
set CGO_ENABLED=0
set GOOS=windows
set GOARCH=amd64
go build -ldflags "-s -w" -o bin\datalogger-windows-amd64.exe cmd\main.go
if %ERRORLEVEL% NEQ 0 (echo Failed Windows x64 build & exit /b 1)

echo [2/7] Building Linux x64...
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=amd64
go build -ldflags "-s -w" -o bin\datalogger-linux-amd64 cmd\main.go
if %ERRORLEVEL% NEQ 0 (echo Failed Linux x64 build & exit /b 1)

echo [3/7] Building Linux ARM64 (Raspberry Pi 4/5 / Orange Pi)...
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=arm64
go build -ldflags "-s -w" -o bin\datalogger-linux-arm64 cmd\main.go
if %ERRORLEVEL% NEQ 0 (echo Failed Linux ARM64 build & exit /b 1)

echo [4/7] Building Linux ARMv7 (Raspberry Pi 2/3 32-bit)...
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=arm
set GOARM=7
go build -ldflags "-s -w" -o bin\datalogger-linux-armv7 cmd\main.go
if %ERRORLEVEL% NEQ 0 (echo Failed Linux ARMv7 build & exit /b 1)

echo [5/7] Building Linux ARMv6 (Raspberry Pi Zero/Zero W / Pi 1)...
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=arm
set GOARM=6
go build -ldflags "-s -w" -o bin\datalogger-linux-armv6 cmd\main.go
if %ERRORLEVEL% NEQ 0 (echo Failed Linux ARMv6 build & exit /b 1)

echo [6/7] Building macOS Intel x64 (Darwin amd64)...
set CGO_ENABLED=0
set GOOS=darwin
set GOARCH=amd64
go build -ldflags "-s -w" -o bin\datalogger-darwin-amd64 cmd\main.go
if %ERRORLEVEL% NEQ 0 (echo Failed macOS Intel build & exit /b 1)

echo [7/7] Building macOS Apple Silicon ARM64 (Darwin arm64)...
set CGO_ENABLED=0
set GOOS=darwin
set GOARCH=arm64
go build -ldflags "-s -w" -o bin\datalogger-darwin-arm64 cmd\main.go
if %ERRORLEVEL% NEQ 0 (echo Failed macOS ARM64 build & exit /b 1)

echo ========================================================
echo Cross-platform build completed successfully!
dir bin
