@echo off
setlocal EnableDelayedExpansion

echo Setting up the Totodoro Backend Project...

REM =============================
REM Resolve project root
REM =============================
pushd "%~dp0\.."
set "ROOT_DIR=%CD%"
set "PROTO_PATH=%ROOT_DIR%\proto"
set "PROTOC_DIR=%ROOT_DIR%\third_party\protobuf"
set "GOOGLEAPIS_DIR=%ROOT_DIR%\third_party\googleapis"
set "PROTO_PACKAGES_DIR=%ROOT_DIR%\internal\proto_package"
set "BIN_DIR=%ROOT_DIR%\bin"
popd

REM =============================
REM Helper checks
REM =============================
where go >nul 2>&1
if errorlevel 1 (
    echo Error: Go is required but not installed or not in PATH.
    exit /b 1
)

where powershell >nul 2>&1
if errorlevel 1 (
    echo Error: PowerShell is required but not found in PATH.
    exit /b 1
)

where git >nul 2>&1
if errorlevel 1 (
    echo Error: Git is required but not installed or not in PATH.
    exit /b 1
)

REM =============================
REM Step 1: Initialize go module
REM =============================
if not exist "%ROOT_DIR%\go.mod" (
    echo Initializing Go module...
    go mod init github.com/latrung124/Totodoro-Backend
) else (
    echo go.mod already exists, skipping module init...
)

REM =============================
REM Step 2: Create bin directory
REM =============================
if not exist "%BIN_DIR%" (
    mkdir "%BIN_DIR%"
    echo Created bin directory
)

REM =============================
REM Step 3: Setup protoc
REM =============================
if not exist "%PROTOC_DIR%\bin\protoc.exe" (
    echo Downloading and extracting Protocol Buffers v32.0-rc1...

    if not exist "%ROOT_DIR%\third_party" mkdir "%ROOT_DIR%\third_party"

    set "ZIP_URL=https://github.com/protocolbuffers/protobuf/releases/download/v32.0-rc1/protoc-32.0-rc-1-win64.zip"
    set "ZIP_FILE=%ROOT_DIR%\protoc-32.0-rc-1-win64.zip"

    powershell -Command "Invoke-WebRequest -Uri '%ZIP_URL%' -OutFile '%ZIP_FILE%'"
    if errorlevel 1 (
        echo Error: Failed to download %ZIP_FILE%
        exit /b 1
    )

    powershell -Command "Expand-Archive -Path '%ZIP_FILE%' -DestinationPath '%PROTOC_DIR%' -Force"
    if errorlevel 1 (
        echo Error: Failed to extract %ZIP_FILE%
        del "%ZIP_FILE%"
        exit /b 1
    )

    del "%ZIP_FILE%"
    if not exist "%PROTOC_DIR%\bin\protoc.exe" (
        echo Error: protoc.exe not found after extraction
        rmdir /S /Q "%PROTOC_DIR%"
        exit /b 1
    )
) else (
    echo Using existing Protocol Buffers in "%PROTOC_DIR%"
)

set "PROTOC=%PROTOC_DIR%\bin\protoc.exe"

REM =============================
REM Step 4: Clone googleapis protos
REM =============================
if not exist "%GOOGLEAPIS_DIR%" (
    echo Cloning googleapis repo...
    git clone https://github.com/googleapis/googleapis.git "%GOOGLEAPIS_DIR%"
) else (
    echo googleapis repo already exists, pulling latest...
    pushd "%GOOGLEAPIS_DIR%"
    git pull --quiet
    popd
)

REM =============================
REM Step 5: Install protoc plugins if missing
REM =============================
set "GOBIN=%GOPATH%\bin"
if "%GOBIN%"=="" for /f "delims=" %%i in ('go env GOPATH') do set "GOBIN=%%i\bin"
set "PATH=%GOBIN%;%PATH%"

where protoc-gen-go >nul 2>&1
if errorlevel 1 (
    echo Installing protoc-gen-go...
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
)

where protoc-gen-go-grpc >nul 2>&1
if errorlevel 1 (
    echo Installing protoc-gen-go-grpc...
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
)

REM NEW: grpc-gateway v2 plugins
where protoc-gen-grpc-gateway >nul 2>&1
if errorlevel 1 (
    echo Installing protoc-gen-grpc-gateway...
    go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
)

where protoc-gen-openapiv2 >nul 2>&1
if errorlevel 1 (
    echo Installing protoc-gen-openapiv2...
    go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
)
# ...existing code...

REM =============================
REM Step 6: Generate Go files from .proto
REM =============================
echo Checking proto path: "%PROTO_PATH%"
if not exist "%PROTO_PATH%" (
    echo Error: Proto path not found: %PROTO_PATH%
    exit /b 1
)

if not exist "%PROTO_PACKAGES_DIR%" (
    mkdir "%PROTO_PACKAGES_DIR%"
    echo Created "%PROTO_PACKAGES_DIR%" directory
)

REM NEW: ensure OpenAPI output dir exists
set "OPENAPI_DIR=%ROOT_DIR%\openapi"
if not exist "%OPENAPI_DIR%" (
    mkdir "%OPENAPI_DIR%"
    echo Created "%OPENAPI_DIR%" directory
)

for /r "%PROTO_PATH%" %%F in (*.proto) do (
    echo Processing %%F...
    set "FILENAME=%%~nF"
    "%PROTOC%" ^
      --go_out=. ^
      --go_opt=paths=source_relative ^
      --go_opt=Mproto/%%~nF.proto=github.com/latrung124/Totodoro-Backend/internal/proto_package/%%~nF ^
      --go-grpc_out=. ^
      --go-grpc_opt=paths=source_relative ^
      --grpc-gateway_out=. ^
      --grpc-gateway_opt=paths=source_relative,generate_unbound_methods=true ^
      --openapiv2_out="%OPENAPI_DIR%" ^
      --openapiv2_opt=logtostderr=true ^
      --proto_path="%PROTO_PATH%" ^
      --proto_path="%PROTOC_DIR%\include" ^
      --proto_path="%GOOGLEAPIS_DIR%" ^
      "%%F"

    if errorlevel 1 (
        echo Error: Failed to generate protobuf files for %%F
        exit /b 1
    ) else (
        echo Successfully generated protobuf files for %%F
    )

    set "TARGET_DIR=%PROTO_PACKAGES_DIR%\%%~nF"
    if not exist "!TARGET_DIR!" mkdir "!TARGET_DIR!"

    if exist "%%~nF.pb.go" move /Y "%%~nF.pb.go" "!TARGET_DIR!" >nul
    if exist "%%~nF_grpc.pb.go" move /Y "%%~nF_grpc.pb.go" "!TARGET_DIR!" >nul
    REM NEW: move grpc-gateway stubs alongside others
    if exist "%%~nF.pb.gw.go" move /Y "%%~nF.pb.gw.go" "!TARGET_DIR!" >nul
)

REM =============================
REM Step 7: Tidy and build project
REM =============================
echo Downloading dependencies...
go mod tidy

echo Building the project...
cd /d "%ROOT_DIR%"
if not exist "main.go" (
    echo Error: main.go not found in the project root
    exit /b 1
)

echo Running go build with verbose output...
go build -v -o "%BIN_DIR%\totodoro-backend.exe" .
if errorlevel 1 (
    echo Error: Build failed
    go build -o "%BIN_DIR%\totodoro-backend.exe" . 2> build_errors.log
    type build_errors.log
    exit /b 1
)

echo Build succeeded. Executable located at "%BIN_DIR%\totodoro-backend.exe"