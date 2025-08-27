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
need_cmd git || die "Git is required. Install Git and re-run."
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

# Paths
PROTO_PATH="$ROOT_DIR/proto"
PROTOC_DIR="$ROOT_DIR/third_party/protobuf"
GOOGLEAPIS_DIR="$ROOT_DIR/third_party/googleapis"
GRPC_GATEWAY_DIR="$ROOT_DIR/third_party/grpc-gateway"
PROTO_PACKAGES_DIR="$ROOT_DIR/internal/proto_package"
OPENAPI_DIR="$ROOT_DIR/openapi"
BIN_DIR="$ROOT_DIR/bin"

# Step 1: Initialize go module (optional if already exists)
if [[ ! -f go.mod ]]; then
  echo "Initializing Go module..."
  go mod init github.com/latrung124/Totodoro-Backend
else
  echo "go.mod already exists, skipping module init..."
fi

# Step 2: Create dirs
mkdir -p "$BIN_DIR" "$PROTO_PACKAGES_DIR" "$OPENAPI_DIR" "$ROOT_DIR/third_party"
echo "Ensured bin, proto_package, openapi, and third_party directories exist"

# Step 3: Setup protoc
PROTOC_BIN="$PROTOC_DIR/bin/protoc"
if [[ ! -x "$PROTOC_BIN" ]]; then
  echo "Downloading and extracting Protocol Buffers v32.0-rc1..."
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

# Step 4: Clone googleapis and grpc-gateway protos
if [[ ! -d "$GOOGLEAPIS_DIR/.git" ]]; then
  echo "Cloning googleapis repo..."
  git clone https://github.com/googleapis/googleapis.git "$GOOGLEAPIS_DIR"
else
  echo "googleapis repo already exists, pulling latest..."
  git -C "$GOOGLEAPIS_DIR" pull --quiet
fi

if [[ ! -d "$GRPC_GATEWAY_DIR/.git" ]]; then
  echo "Cloning grpc-gateway repo..."
  git clone https://github.com/grpc-ecosystem/grpc-gateway.git "$GRPC_GATEWAY_DIR"
else
  echo "grpc-gateway repo already exists, pulling latest..."
  git -C "$GRPC_GATEWAY_DIR" pull --quiet
fi

# Step 5: Install protoc plugins
export PATH="$(go env GOPATH)/bin:$PATH"
if ! command -v protoc-gen-go >/dev/null 2>&1; then
  echo "Installing protoc-gen-go..."
  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
fi
if ! command -v protoc-gen-go-grpc >/dev/null 2>&1; then
  echo "Installing protoc-gen-go-grpc..."
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
fi
# grpc-gateway v2 plugins
if ! command -v protoc-gen-grpc-gateway >/dev/null 2>&1; then
  echo "Installing protoc-gen-grpc-gateway..."
  go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
fi
if ! command -v protoc-gen-openapiv2 >/dev/null 2>&1; then
  echo "Installing protoc-gen-openapiv2..."
  go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
fi

# Step 6: Generate Go files from all .proto files
echo "Checking proto path: \"$PROTO_PATH\""
[[ -d "$PROTO_PATH" ]] || die "Proto path not found: $PROTO_PATH"

echo "Generating protobuf, gRPC, gateway, and OpenAPI files..."
while IFS= read -r -d '' proto; do
  echo "Processing $proto..."
  base="$(basename "$proto" .proto)"

  "$PROTOC_BIN" \
    --go_out=. \
    --go_opt=paths=source_relative \
    --go_opt=Mproto/${base}.proto=github.com/latrung124/Totodoro-Backend/internal/proto_package/${base} \
    --go-grpc_out=. \
    --go-grpc_opt=paths=source_relative \
    --grpc-gateway_out=. \
    --grpc-gateway_opt=paths=source_relative,generate_unbound_methods=true \
    --openapiv2_out="$OPENAPI_DIR" \
    --openapiv2_opt=logtostderr=true \
    --proto_path="$PROTO_PATH" \
    --proto_path="$PROTOC_DIR/include" \
    --proto_path="$GOOGLEAPIS_DIR" \
    --proto_path="$GRPC_GATEWAY_DIR" \
    "$proto"

  target_dir="$PROTO_PACKAGES_DIR/$base"
  mkdir -p "$target_dir"

  # Move generated files if they appear in CWD or proto/
  for f in "${base}.pb.go" "${base}_grpc.pb.go" "${base}.pb.gw.go"; do
    if [[ -f "$f" ]]; then
      mv -f "$f" "$target_dir"/
    elif [[ -f "proto/$f" ]]; then
      mv -f "proto/$f" "$target_dir"/
    fi
  done
  echo "Generated into $target_dir"
done < <(find "$PROTO_PATH" -type f -name '*.proto' -print0)

# Step 7: Now that proto packages exist, tidy modules
echo "Downloading dependencies (go mod tidy)..."
go mod tidy

# Step 8: Build the project
echo "Building the project..."
[[ -f "main.go" ]] || die "main.go not found in the project root"
if ! go build -v -o "$BIN_DIR/totodoro-backend" .; then
  echo "Error: Build failed"
  go build -o "$BIN_DIR/totodoro-backend" . 2> build_errors.log || true
  echo "---------- build_errors.log ----------"
  cat build_errors.log || true
  exit 1
fi

echo "✅ Setup complete. Run your server using: ./bin/totodoro-backend"