@echo off
REM Build script for Windows
SET GOOS=windows
SET GOARCH=amd64

cd /d "%~dp0"  REM Ensure we are in the project root

if not exist "main.go" (
    echo Error: main.go not found in the project root
    exit /b 1
)

echo Running go build with verbose output...

go build -v -o bin\totodoro-backend.exe . || (
    echo Error: Build failed
    echo Capturing build errors...
    go build -o bin\totodoro-backend.exe . 2> build_errors.log
    type build_errors.log
    exit /b 1
)

echo ✅ Setup complete. Run your server using: bin\totodoro-backend.exe
exit /b 0