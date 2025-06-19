#!/usr/bin/env bash
set -euo pipefail

: "${OPENFHE_ROOT:="/usr/local"}"
: "${EMCC:=$(command -v emcc || true)}"
if [[ -z "$EMCC" ]]; then
  echo "Error: emcc not found. Please source your emsdk_env.sh" >&2
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SRC="$SCRIPT_DIR/src/openfhe_c.cpp"
OUT="$SCRIPT_DIR/../web/wasm"
mkdir -p "$OUT"

# === include paths ===
CFLAGS=(
  -I"$SCRIPT_DIR/src"
  -I"$OPENFHE_ROOT/include"                           # so <openfhe/...> and <cereal/...> work
  -I"$OPENFHE_ROOT/include/openfhe"                   # parent of core/, pke/, binfhe/, cereal/
  -I"$OPENFHE_ROOT/include/openfhe/core"
  -I"$OPENFHE_ROOT/include/openfhe/core/lattice"
  -I"$OPENFHE_ROOT/include/openfhe/pke"
  -I"$OPENFHE_ROOT/include/openfhe/binfhe"             # <<–– for binfhecontext.h
  -I"$OPENFHE_ROOT/include/openfhe/cereal"
)

COMMON_FLAGS=(
  -O3
  -std=c++17
  -s STANDALONE_WASM=1
  -s ALLOW_MEMORY_GROWTH=1
)

LIBS="-L$OPENFHE_ROOT/lib -lopenfhe"

echo "▶ Building openfhe_keygen.wasm…"
$EMCC "${SRC}" "${CFLAGS[@]}" "${COMMON_FLAGS[@]}" $LIBS \
  -s EXPORTED_FUNCTIONS='["_go_bgvrns_new","_go_bgvrns_free","_go_keygen_ptr","_go_pk_serialize","_go_buf_free"]' \
  -o "$OUT/openfhe_keygen.wasm"

echo "▶ Building openfhe_encrypt.wasm…"
$EMCC "${SRC}" "${CFLAGS[@]}" "${COMMON_FLAGS[@]}" $LIBS \
  -s EXPORTED_FUNCTIONS='["_go_bgvrns_new","_go_bgvrns_free","_go_pk_deserialize","_go_encrypt_u64_ptr_out","_go_ct_ser","_go_buf_free"]' \
  -o "$OUT/openfhe_encrypt.wasm"

echo "▶ Building openfhe_decrypt.wasm…"
$EMCC "${SRC}" "${CFLAGS[@]}" "${COMMON_FLAGS[@]}" $LIBS \
  -s EXPORTED_FUNCTIONS='["_go_bgvrns_new","_go_bgvrns_free","_go_sk_deserialize","_go_decrypt_u64_ptr","_go_buf_free"]' \
  -o "$OUT/openfhe_decrypt.wasm"

echo "✅ Modules written to $OUT"
