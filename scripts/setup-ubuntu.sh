#!/usr/bin/env bash
set -euo pipefail
shopt -s nullglob
[[ "${DEBUG:-}" == "1" ]] && set -x
set -o errtrace

err() {
  local exit_code=$?
  echo "Error at line ${BASH_LINENO[0]}: ${BASH_COMMAND}" >&2
  exit "$exit_code"
}
trap err ERR

echo "Setting up the Totodoro Backend Project..."

# Resolve project root (parent of scripts/)
SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$ROOT_DIR"

need_cmd() { command -v "$1" >/dev/null 2>&1; }
die() { echo "Error: $*" >&2; exit 1; }

# Basic tool checks
need_cmd go || die "Go is required. Install Go and re-run."
if ! need_cmd unzip; then
  echo "Installing unzip..."
  if need_cmd apt-get; then
    sudo apt-get update -y && sudo apt-get install -y unzip
  else
    die "unzip not found. Please install unzip."
  fi
fi

download() {
  # usage: download <output-file> <url>
  local out="$1" url="$2"
  if command -v curl >/dev/null 2>&1; then
    curl -fL --retry 3 --connect-timeout 15 -o "$out" "$url"
  elif command -v wget >/dev/null 2>&1; then
    wget -qO "$out" "$url"
  else
    die "Neither curl nor wget found. Please install one of them."
  fi
}

echo "Go version: $(go version)"

# Step 1: Initialize go module (optional if already exists)
if [[ ! -f go.mod ]]; then
  echo "Initializing Go module..."
  go mod init github.com/latrung124/Totodoro-Backend
else
  echo "go.mod already exists, skipping module init..."
fi

# Step 3: Create bin directory
mkdir -p bin
echo "Ensured bin directory exists"

# Step 4: Check and setup Protocol Buffers v32.0-rc1
PROTOC_DIR="third_party/protobuf"
PROTOC_BIN="$PROTOC_DIR/bin/protoc"
if [[ ! -x "$PROTOC_BIN" ]]; then
  echo "Downloading and extracting Protocol Buffers v32.0-rc1..."
  mkdir -p third_party
  ZIP_URL="https://github.com/protocolbuffers/protobuf/releases/download/v32.0-rc1/protoc-32.0-rc-1-linux-x86_64.zip"
  ZIP_FILE="protoc-32.0-rc-1-linux-x86_64.zip"
  download "$ZIP_FILE" "$ZIP_URL" || die "Failed to download $ZIP_FILE"

  rm -rf "$PROTOC_DIR"
  mkdir -p "$PROTOC_DIR"
  unzip -o "$ZIP_FILE" -d "$PROTOC_DIR"
  rm -f "$ZIP_FILE"

  [[ -x "$PROTOC_BIN" ]] || { rm -rf "$PROTOC_DIR"; die "protoc not found after extraction"; }
else
  echo "Using existing Protocol Buffers in $PROTOC_DIR"
fi

echo "protoc version: $("$PROTOC_BIN" --version)"

# Ensure protoc plugins are available
export PATH="$(go env GOPATH)/bin:$PATH"
if ! command -v protoc-gen-go >/dev/null 2>&1; then
  echo "Installing protoc-gen-go..."
  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
fi
if ! command -v protoc-gen-go-grpc >/dev/null 2>&1; then
  echo "Installing protoc-gen-go-grpc..."
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
fi

# Step 5: Generate Go files from all .proto files
PROTO_PATH="$ROOT_DIR/proto"
echo "Checking proto path: \"$PROTO_PATH\""
[[ -d "$PROTO_PATH" ]] || die "Proto path not found: $PROTO_PATH"

echo "Current working directory: \"$PWD\""
echo "Generating protobuf files..."

PROTO_PACKAGES_DIR="$ROOT_DIR/internal/proto_package"
mkdir -p "$PROTO_PACKAGES_DIR"
echo "Ensured \"$PROTO_PACKAGES_DIR\" directory exists"

while IFS= read -r -d '' proto; do
  echo "Processing $proto..."
  base="$(basename "$proto" .proto)"

  "$PROTOC_BIN" \
    --go_out=. \
    --go_opt=paths=source_relative \
    --go_opt=Mproto/${base}.proto=github.com/latrung124/Totodoro-Backend/internal/proto_package/${base} \
    --go-grpc_out=. \
    --go-grpc_opt=paths=source_relative \
    --proto_path="$PROTO_PATH" \
    --proto_path="$PROTOC_DIR/include" \
    "$proto"

  echo "Successfully generated protobuf files for $proto"

  target_dir="$PROTO_PACKAGES_DIR/$base"
  mkdir -p "$target_dir"

  pb_file="${base}.pb.go"
  grpc_file="${base}_grpc.pb.go"

  if [[ -f "$pb_file" ]]; then
    mv -f "$pb_file" "$target_dir"/
  elif [[ -f "proto/$pb_file" ]]; then
    mv -f "proto/$pb_file" "$target_dir"/
  else
    echo "Warning: Generated file \"$pb_file\" not found"
  fi

  if [[ -f "$grpc_file" ]]; then
    mv -f "$grpc_file" "$target_dir"/
  elif [[ -f "proto/$grpc_file" ]]; then
    mv -f "proto/$grpc_file" "$target_dir"/
  else
    echo "Warning: Generated file \"$grpc_file\" not found"
  fi
done < <(find "$PROTO_PATH" -type f -name '*.proto' -print0)

# Step 6: Now that proto packages exist, tidy modules
echo "Downloading dependencies..."
go mod tidy

# Step 7: Build the project
echo "Building the project..."
[[ -f "main.go" ]] || die "main.go not found in the project root"

echo "Running go build with verbose output..."
if ! go build -v -o bin/totodoro-backend .; then
  echo "Error: Build failed"
  echo "Capturing build errors..."
  go build -o bin/totodoro-backend . 2> build_errors.log || true
  cat build_errors.log || true
  exit 1
fi

echo "✅ Setup complete. Run your server using: ./bin/totodoro-backend"