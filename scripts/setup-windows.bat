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
set "GRPC_GATEWAY_DIR=%ROOT_DIR%\third_party\grpc-gateway"
REM For protoc includes from grpc-gateway (openapiv2 options)
set "PROTO_PACKAGES_DIR=%ROOT_DIR%\internal\proto_package"
set "BIN_DIR=%ROOT_DIR%\bin"
popd

REM ...existing code for checks, module init, bin dir ...

REM Step 3: Setup protoc
REM ...existing code...

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
REM Step 4b: Clone grpc-gateway (for openapiv2 proto includes)
REM =============================
if not exist "%GRPC_GATEWAY_DIR%" (
    echo Cloning grpc-gateway repo...
    git clone https://github.com/grpc-ecosystem/grpc-gateway.git "%GRPC_GATEWAY_DIR%"
) else (
    echo grpc-gateway repo already exists, pulling latest...
    pushd "%GRPC_GATEWAY_DIR%"
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

REM ensure OpenAPI output dir exists
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
      --proto_path="%GRPC_GATEWAY_DIR%" ^
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