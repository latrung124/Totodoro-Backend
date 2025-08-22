#!/usr/bin/env bash
set -euo pipefail
[[ "${DEBUG:-}" == "1" ]] && set -x
set -o errtrace

err() {
  local exit_code=$?
  echo "Error at line ${BASH_LINENO[0]}: ${BASH_COMMAND}" >&2
  exit "$exit_code"
}
trap err ERR

# Resolve project root (parent of scripts/)
SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$ROOT_DIR"

command -v go >/dev/null 2>&1 || { echo "Error: Go is required"; exit 1; }

# Optional: ensure generated protobuf code exists
if [[ ! -d "internal/proto_package" ]] || ! compgen -G "internal/proto_package/*/*.go" >/dev/null; then
  echo "Warning: Generated protobuf code not found in internal/proto_package/*"
  echo "If build fails, run: scripts/setup-ubuntu.sh"
fi

mkdir -p bin

GOOS="${GOOS:-$(go env GOOS)}"
GOARCH="${GOARCH:-$(go env GOARCH)}"
BIN_NAME="totodoro-backend"
[[ "$GOOS" == "windows" ]] && BIN_NAME="${BIN_NAME}.exe"

OUT="bin/${BIN_NAME}"

echo "Building for ${GOOS}/${GOARCH} -> ${OUT}"
echo "Go version: $(go version)"

# Pass any additional args to go build, e.g. ./scripts/build-ubuntu.sh -race -tags prod
if ! go build -v -o "$OUT" "$@" .; then
  echo "Build failed, capturing errors to build_errors.log"
  go build -o "$OUT" "$@" . 2> build_errors.log || true
  echo "---------- build_errors.log ----------"
  cat build_errors.log || true
  exit 1
fi

echo "Build succeeded: $OUT"