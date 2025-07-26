@echo off
echo Setting up the Totodoro Backend Project...

REM Step 1: Initialize go module (optional if already exists)
IF NOT EXIST "go.mod" (
    echo Initializing Go module...
    go mod init github.com/latrung124/Totodoro-Backend
) ELSE (
    echo go.mod already exists, skipping module init...
)

REM Step 2: Install dependencies
echo Downloading dependencies...
go mod tidy

REM Step 3: Create .env file if it doesn't exist
IF NOT EXIST ".env" (
    echo Creating default .env file...
    echo GIN_MODE=debug > .env
    echo PORT=8080 >> .env
) ELSE (
    echo .env already exists, skipping creation...
)

REM Step 4: Create bin directory
IF NOT EXIST "bin" (
    mkdir bin
    echo Created bin directory
)

REM Step 5: Optional - Build the project
echo Building the project...
SET GOOS=windows
SET GOARCH=amd64
go build -o bin\totodoro-backend.exe main.go

echo ✅ Setup complete. Run your server using: bin\\totodoro-backend.exe
