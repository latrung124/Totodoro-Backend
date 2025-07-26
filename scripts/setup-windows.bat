@echo off
setlocal EnableDelayedExpansion

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

REM Step 3: Create bin directory
IF NOT EXIST "bin" (
    mkdir bin
    echo Created bin directory
)

REM Step 4: Check and setup Protocol Buffers v32.0-rc1
set "PROTOC_DIR=third_party/protobuf"
if not exist "%PROTOC_DIR%\bin\protoc.exe" (
    echo Downloading and extracting Protocol Buffers v32.0-rc1...
    
    :: Create third_party directory if it doesn't exist
    if not exist third_party mkdir third_party
    
    :: Download protoc-32.0-rc-1-win64.zip
    set "ZIP_URL=https://github.com/protocolbuffers/protobuf/releases/download/v32.0-rc1/protoc-32.0-rc-1-win64.zip"
    set "ZIP_FILE=protoc-32.0-rc-1-win64.zip"
    powershell -Command "Invoke-WebRequest -Uri !ZIP_URL! -OutFile !ZIP_FILE!"
    if !ERRORLEVEL! neq 0 (
        echo Error: Failed to download !ZIP_FILE!
        exit /b 1
    )

    :: Extract the ZIP file
    powershell -Command "Expand-Archive -Path !ZIP_FILE! -DestinationPath !PROTOC_DIR! -Force"
    if !ERRORLEVEL! neq 0 (
        echo Error: Failed to extract !ZIP_FILE!
        del !ZIP_FILE!
        exit /b 1
    )

    :: Clean up the ZIP file
    del !ZIP_FILE!
    
    :: Verify extraction
    if not exist "%PROTOC_DIR%\bin\protoc.exe" (
        echo Error: protoc.exe not found after extraction
        rmdir /S /Q !PROTOC_DIR!
        exit /b 1
    )
) else (
    echo Using existing Protocol Buffers in !PROTOC_DIR!
)

REM Step 6: Generate Go files from all .proto files in the proto folder using the extracted protoc
echo Generating protobuf files...
set "PROTOC=%PROTOC_DIR%\bin\protoc.exe"

REM Set and enable delayed expansion for PROTO_PATH
set "PROTO_PATH=proto"

REM Change to the project root directory to ensure relative paths work
cd /d "%~dp0"

REM Iterate over all .proto files in the proto folder and its subdirectories
for /r !PROTO_PATH! %%F in (*.proto) do (
    echo Processing %%F...
    set "REL_PATH=%%F"
    set "REL_PATH=!REL_PATH:%CD%\=!"
    echo Relative path for protoc: !REL_PATH!
    "!PROTOC!" --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative --proto_path=!PROTO_PATH! --proto_path=%PROTOC_DIR%\include "!REL_PATH!"
    if !ERRORLEVEL! neq 0 (
        echo Error: Failed to generate protobuf files for !REL_PATH!
        exit /b 1
    )
)

REM Move generated files to their respective package directories (e.g., internal/<package>)
for /r . %%G in (*.pb.go *.grpc.pb.go) do (
    if exist %%G (
        for /f "delims=" %%H in ("%%~dpG") do set "SOURCE_DIR=%%~nxH"
        if "!SOURCE_DIR!"=="proto" (
            move "%%G" internal\
        ) else (
            move "%%G" internal\"!SOURCE_DIR!"
        )
        if !ERRORLEVEL! neq 0 (
            echo Error: Failed to move generated file %%G
            exit /b 1
        )
    )
)

if !ERRORLEVEL! neq 0 (
    echo Error: Failed to generate protobuf files
    exit /b 1
)

endlocal

REM Step 7: Build the project
echo Building the project...
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
