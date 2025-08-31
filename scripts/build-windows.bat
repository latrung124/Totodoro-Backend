@echo off
setlocal EnableDelayedExpansion

REM Resolve project root and paths
pushd "%~dp0\.."
set "ROOT_DIR=%CD%"
set "BIN_DIR=%ROOT_DIR%\bin"
set "OUT=%BIN_DIR%\totodoro-backend.exe"

REM Ensure bin directory exists
if not exist "%BIN_DIR%" (
    mkdir "%BIN_DIR%"
)

REM Check Go availability
where go >nul 2>&1
if errorlevel 1 (
    echo Error: Go toolchain not found on PATH
    popd
    exit /b 1
)

REM Build for Windows amd64
set "GOOS=windows"
set "GOARCH=amd64"

REM Optional: ensure dependencies are tidy
echo Downloading dependencies...
go mod tidy
if errorlevel 1 (
    echo Error: go mod tidy failed
    popd
    exit /b 1
)

REM Verify main.go at repo root
if not exist "%ROOT_DIR%\main.go" (
    echo Error: main.go not found in the project root
    popd
    exit /b 1
)

echo Running go build with verbose output...
if exist "%ROOT_DIR%\build_errors.log" del /q "%ROOT_DIR%\build_errors.log"

REM Pass through any extra flags: scripts\build-windows.bat -- -race -tags prod
set "EXTRA_ARGS="
if "%1"=="--" (
    shift
)
:collect_args
if "%~1"=="" goto do_build
set "EXTRA_ARGS=%EXTRA_ARGS% %~1"
shift
goto collect_args

:do_build
go build -v -o "%OUT%" %EXTRA_ARGS% .
if errorlevel 1 (
    echo Error: Build failed
    echo Capturing build errors...
    go build -o "%OUT%" %EXTRA_ARGS% . 2> "%ROOT_DIR%\build_errors.log"
    type "%ROOT_DIR%\build_errors.log"
    popd
    exit /b 1
)

echo Build succeeded. Executable located at "%OUT%"
popd
exit /b 0