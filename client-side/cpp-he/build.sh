#!/usr/bin/env bash
set -euo pipefail

: "${OPENFHE_ROOT:=/path/to/emscripten_build/install}"
: "${EMCC:=$(command -v emcc || true)}"
[[ -x "$EMCC" ]] || { echo "Error: emcc not found" >&2; exit 1; }

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SRC="$SCRIPT_DIR/src/openfhe_c.cpp"
OUT="$SCRIPT_DIR/../web/wasm"
mkdir -p "$OUT"

# include paths
CFLAGS=(
  -I"$SCRIPT_DIR/src"
  -I"$OPENFHE_ROOT/include"
  -I"$OPENFHE_ROOT/include/openfhe"
  -I"$OPENFHE_ROOT/include/openfhe/core"
  -I"$OPENFHE_ROOT/include/openfhe/core/lattice"
  -I"$OPENFHE_ROOT/include/openfhe/pke"
  -I"$OPENFHE_ROOT/include/openfhe/binfhe"
  -I"$OPENFHE_ROOT/include/openfhe/cereal"
)

COMMON_FLAGS=(
  -O3
  -std=c++17
  -s STANDALONE_WASM=1       # standalone module
  -s ALLOW_MEMORY_GROWTH=1
)

# point at each static archive
LIBS=(
  "$OPENFHE_ROOT/lib/libOPENFHEcore_static.a"
  "$OPENFHE_ROOT/lib/libOPENFHEpke_static.a"
  "$OPENFHE_ROOT/lib/libOPENFHEbinfhe_static.a"
)

# wrap them in --whole-archive to force-link every symbol
LDFLAGS=(
  -Wl,--whole-archive
    "${LIBS[@]}"
  -Wl,--no-whole-archive
  -Wl,--no-entry           # no 'main' needed
)

# helper to build each module
build_module(){
  local name=$1 exports=$2
  echo "▶ Building $name…"
  $EMCC \
    "$SRC" "${CFLAGS[@]}" "${COMMON_FLAGS[@]}" \
    "${LDFLAGS[@]}" \
    -s EXPORTED_FUNCTIONS="[$exports]" \
    -o "$OUT/$name"
}

build_module openfhe_keygen.wasm '"_go_bgvrns_new","_go_bgvrns_free","_go_keygen_ptr","_go_pk_serialize","_go_buf_free"'
build_module openfhe_encrypt.wasm '"_go_bgvrns_new","_go_bgvrns_free","_go_pk_deserialize","_go_encrypt_u64_ptr_out","_go_ct_ser","_go_buf_free"'
build_module openfhe_decrypt.wasm '"_go_bgvrns_new","_go_bgvrns_free","_go_sk_deserialize","_go_decrypt_u64_ptr","_go_buf_free"'

echo "✅ Modules written to $OUT"
