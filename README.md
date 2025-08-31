# Totodoro-Backend

Microservices Monolith for Totodoro-Backend.

## Quick Start

- Windows
  1. Open a terminal in the repo root.
  2. Run setup (generates protobufs and builds):
     ```bat
     scripts\setup-windows.bat
     ```
  3. Run the server:
     ```bat
     bin\totodoro-backend.exe
     ```

- Ubuntu
  1. Make scripts executable:
     ```bash
     chmod +x scripts/setup-ubuntu.sh scripts/build-ubuntu.sh
     ```
  2. Run setup (downloads protoc, installs plugins, generates protobufs):
     ```bash
     ./scripts/setup-ubuntu.sh
     ```
  3. Build:
     ```bash
     ./scripts/build-ubuntu.sh
     ```
  4. Run the server:
     ```bash
     ./bin/totodoro-backend
     ```

## Prerequisites

- Go 1.20+ (https://go.dev/dl/)
- Git

Platform-specific:
- Windows: PowerShell available (default). The setup script downloads protoc automatically.
- Ubuntu: curl or wget, unzip (the setup script installs unzip via apt if missing).

Protocol Buffers:
- Both setup scripts install protoc (v32.0-rc1) locally to third_party/protobuf and install Go plugins:
  - google.golang.org/protobuf/cmd/protoc-gen-go
  - google.golang.org/grpc/cmd/protoc-gen-go-grpc

## Detailed Setup

- Windows
  - Run:
    ```bat
    scripts\setup-windows.bat
    ```
  - This will:
    - Initialize go.mod if missing
    - Download dependencies
    - Download and extract protoc (if not present)
    - Generate Go files from proto/*.proto into internal/proto_package/*
    - Build bin\totodoro-backend.exe

- Ubuntu
  - Run:
    ```bash
    ./scripts/setup-ubuntu.sh
    ```
  - This will:
    - Initialize go.mod if missing
    - Ensure unzip is installed
    - Download and extract protoc (if not present)
    - Install protoc-gen-go and protoc-gen-go-grpc
    - Generate Go files from proto/*.proto into internal/proto_package/*
    - Tidy Go modules

## Build Only

- Cross-platform (Python wrapper)
  - Ensure Python 3 is installed.
  - The wrapper auto-detects your OS and dispatches to the correct script:
    - Windows -> scripts\build-windows.bat / scripts\setup-windows.bat
    - Linux/macOS -> scripts/build-ubuntu.sh / scripts/setup-ubuntu.sh
  ```bash
  # Setup (generate protobufs, install tools)
  python3 scripts/run.py setup
  # Windows alternative:
  py scripts\run.py setup

  # Build only
  python3 scripts/run.py build
  # Windows alternative:
  py scripts\run.py build

  # Pass extra flags to go build after a -- separator
  python3 scripts/run.py build -- -race -tags prod
  ```

- Windows
  - If you already ran setup, the binary is built. To rebuild manually:
    ```bat
    go build -v -o bin\totodoro-backend.exe .
    ```

- Ubuntu
  - Run:
    ```bash
    ./scripts/build-ubuntu.sh
    ```
  - Or manually:
    ```bash
    go build -v -o bin/totodoro-backend .
    ```

## Run

- Windows:
  ```bat
  bin\totodoro-backend.exe
  ```
- Ubuntu:
  ```bash
  ./bin/totodoro-backend
  ```

## Generating Protobufs Manually

Both setup scripts already generate protobufs. To regenerate manually:

- Windows:
  ```bat
  scripts\setup-windows.bat
  ```
- Ubuntu:
  ```bash
  ./scripts/setup-ubuntu.sh
  ```

## Tests

Run unit tests (both platforms):
```bash
go test ./...
```

## Troubleshooting

- go mod tidy errors like “no matching versions for query "latest"”:
  - Ensure protobufs are generated first (run the setup script).
- Missing protoc-gen-go or protoc-gen-go-grpc:
  ```bash
  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
  ```
  Ensure $(go env GOPATH)/bin is in PATH.
- Ubuntu verbose logs:
  ```bash
  DEBUG=1 ./scripts/setup-ubuntu.sh
  ```
- Force re-download protoc:
  - Delete third_party/protobuf and re-run setup.