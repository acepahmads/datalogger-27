@echo off
REM ========================================================
REM  Datalogger Cross-Platform Pure-Go Multi-Target Builder
REM  Requires NO CGO compiler (zero CGO pure Go SQLite)
REM ========================================================

echo Compiling cross-platform binaries into bin/...
if not exist "bin" mkdir "bin"

echo [1/4] Building Windows x64...
set CGO_ENABLED=0
set GOOS=windows
set GOARCH=amd64
go build -ldflags "-s -w" -o bin\datalogger-windows-amd64.exe cmd\main.go
if %ERRORLEVEL% NEQ 0 (echo Failed Windows x64 build & exit /b 1)

echo [2/4] Building Linux x64...
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=amd64
go build -ldflags "-s -w" -o bin\datalogger-linux-amd64 cmd\main.go
if %ERRORLEVEL% NEQ 0 (echo Failed Linux x64 build & exit /b 1)

echo [3/4] Building Linux ARM64 (Raspberry Pi 4/5)...
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=arm64
go build -ldflags "-s -w" -o bin\datalogger-linux-arm64 cmd\main.go
if %ERRORLEVEL% NEQ 0 (echo Failed Linux ARM64 build & exit /b 1)

echo [4/4] Building Linux ARM32 (Raspberry Pi Zero/3)...
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=arm
set GOARM=7
go build -ldflags "-s -w" -o bin\datalogger-linux-armv7 cmd\main.go
if %ERRORLEVEL% NEQ 0 (echo Failed Linux ARM32 build & exit /b 1)

echo ========================================================
echo Cross-platform build completed successfully!
dir bin
