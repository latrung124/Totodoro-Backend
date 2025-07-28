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
set "PROTOC_DIR=third_party/protobuf"
set "PROTOC=%PROTOC_DIR%\bin\protoc.exe"
set "PROTO_PATH=%CD%\proto"

echo Checking proto path: "%PROTO_PATH%"

if not exist "%PROTOC%" (
    echo Error: protoc.exe not found in %PROTOC_DIR%
    exit /b 1
)

echo Current working directory: "%CD%"

echo Generating protobuf files...

set "PROTO_PACKAGES_DIR=%CD%\internal\proto_package"
if not exist "%PROTO_PACKAGES_DIR%" (
    mkdir "%PROTO_PACKAGES_DIR%"
    echo Created "%PROTO_PACKAGES_DIR%" directory
)

for /r "%PROTO_PATH%" %%F in (*.proto) do (
    echo Processing %%F...
    REM Get file name without extension
    set "FILENAME=%%~nF"
    "%PROTOC%" --go_out=. --go_opt=paths=source_relative --go_opt=Mproto/%%~nF.proto=github.com/latrung124/Totodoro-Backend/internal/%%~nF --go-grpc_out=. --go-grpc_opt=paths=source_relative --proto_path="%PROTO_PATH%" --proto_path="%PROTOC_DIR%\include" "%%F"
    if !ERRORLEVEL! neq 0 (
        echo Error: Failed to generate protobuf files for %%F
        exit /b 1
    ) else (
        echo Successfully generated protobuf files for %%F
    )

    REM Set the target directory for the generated files
    set "PROTO_PACKAGE_DIR=%CD%\internal\proto_package\%%~nF"
    if not exist "!PROTO_PACKAGE_DIR!" (
        mkdir "!PROTO_PACKAGE_DIR!"
        echo Created "!PROTO_PACKAGE_DIR!" directory
    ) else (
        echo "!PROTO_PACKAGE_DIR!" already exists, skipping creation.
    )

    REM Move only the files generated for the current .proto file
    set "PB_FILE=%%~nF.pb.go"
    set "GRPC_FILE=%%~nF_grpc.pb.go"

    if exist "!PB_FILE!" (
        echo Moving generated file "!PB_FILE!" to "!PROTO_PACKAGE_DIR!"
        move "!PB_FILE!" "!PROTO_PACKAGE_DIR!"
        if !ERRORLEVEL! neq 0 (
            echo Error: Failed to move generated file "!PB_FILE!"
            exit /b 1
        )
    ) else (
        echo Warning: Generated file "!PB_FILE!" not found
    )

    if exist "!GRPC_FILE!" (
        echo Moving generated file "!GRPC_FILE!" to "!PROTO_PACKAGE_DIR!"
        move "!GRPC_FILE!" "!PROTO_PACKAGE_DIR!"
        if !ERRORLEVEL! neq 0 (
            echo Error: Failed to move generated file "!GRPC_FILE!"
            exit /b 1
        )
    ) else (
        echo Warning: Generated file "!GRPC_FILE!" not found
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
