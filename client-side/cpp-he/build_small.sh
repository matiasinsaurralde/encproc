#!/usr/bin/env bash
set -euo pipefail

# ——— CONFIG ———
: "${OPENFHE_ROOT:=/path/to/emscripten_build/install}"
EMCC=${EMCC:-$(command -v emcc)}
[[ -x "$EMCC" ]] || { echo "Error: emcc not found" >&2; exit 1; }

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SRC="$SCRIPT_DIR/src/openfhe_c.cpp"
OUT="$SCRIPT_DIR/../web/wasm"
mkdir -p "$OUT"

# ——— COMMON FLAGS ———
CFLAGS=(
  -I"$SCRIPT_DIR/src"
  -I"$OPENFHE_ROOT/include"
  -I"$OPENFHE_ROOT/include/openfhe"
  -I"$OPENFHE_ROOT/include/openfhe/core"
  -I"$OPENFHE_ROOT/include/openfhe/core/lattice"
  -I"$OPENFHE_ROOT/include/openfhe/pke"
  -I"$OPENFHE_ROOT/include/openfhe/binfhe"
  -I"$OPENFHE_ROOT/include/openfhe/cereal"
  -Oz
  -flto
  -s STANDALONE_WASM=1
  -s ALLOW_MEMORY_GROWTH=1
  -std=c++17
)

# pack all three libs
LIBS=(
  "$OPENFHE_ROOT/lib/libOPENFHEcore_static.a"
  "$OPENFHE_ROOT/lib/libOPENFHEpke_static.a"
  "$OPENFHE_ROOT/lib/libOPENFHEbinfhe_static.a"
)
LDFLAGS=(
  -flto
  -Wl,--whole-archive "${LIBS[@]}" -Wl,--no-whole-archive
  -Wl,--no-entry
)

# ——— BUILD KEYGEN ———
echo "▶ Building openfhe_keygen.wasm…"
"$EMCC" "$SRC" "${CFLAGS[@]}" "${LDFLAGS[@]}" \
  -s EXPORTED_FUNCTIONS='["_go_bgvrns_new","_go_bgvrns_free","_go_keygen_ptr","_go_pk_serialize","_go_buf_free"]' \
  -o "$OUT/openfhe_keygen.wasm"

# ——— BUILD ENCRYPT ———
echo "▶ Building openfhe_encrypt.wasm…"
"$EMCC" "$SRC" "${CFLAGS[@]}" "${LDFLAGS[@]}" \
  -s EXPORTED_FUNCTIONS='["_go_bgvrns_new","_go_bgvrns_free","_go_pk_deserialize","_go_encrypt_u64_ptr_out","_go_ct_ser","_go_buf_free"]' \
  -o "$OUT/openfhe_encrypt.wasm"

# ——— BUILD DECRYPT ———
echo "▶ Building openfhe_decrypt.wasm…"
"$EMCC" "$SRC" "${CFLAGS[@]}" "${LDFLAGS[@]}" \
  -s EXPORTED_FUNCTIONS='["_go_bgvrns_new","_go_bgvrns_free","_go_sk_deserialize","_go_decrypt_u64_ptr","_go_buf_free"]' \
  -o "$OUT/openfhe_decrypt.wasm"

echo "✅ Minified modules written to $OUT"
