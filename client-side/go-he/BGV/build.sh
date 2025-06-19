#!/usr/bin/env bash
set -e

# ensure WASM target
export GOOS=js
export GOARCH=wasm

# output directory
OUTDIR="$(dirname "$0")/../web/go-wasm"

# build each command
go build -o "$OUTDIR/go-genkey.wasm"    ./genkey
go build -o "$OUTDIR/go-encrypt.wasm"   ./encrypt
go build -o "$OUTDIR/go-decrypt.wasm"   ./decrypt

echo "Built Go-WASM modules into $OUTDIR"