// ofh_keygen.js
// Initializes OpenFHE-WASM key generation and exposes generateKeypair()
export async function initOFHKeygen(wasmUrl, depth = 1, t = 0x10001) {
  const resp = await fetch(wasmUrl);
  const { instance } = await WebAssembly.instantiateStreaming(resp, {});
  const { memory, malloc, free, go_bgvrns_new, go_keygen_ptr, go_buf_free } = instance.exports;
  const dv = new DataView(memory.buffer);
  const ptrSize = 4;

  // Create context
  const ctx = go_bgvrns_new(depth, t);

  return {
    generateKeypair: () => {
      const pkPtrPtr = malloc(ptrSize);
      const skPtrPtr = malloc(ptrSize);
      go_keygen_ptr(ctx, pkPtrPtr, skPtrPtr);

      const pkPtr = dv.getUint32(pkPtrPtr, true);
      const skPtr = dv.getUint32(skPtrPtr, true);
      free(pkPtrPtr);
      free(skPtrPtr);

      // Serialize helper
      function serialize(ptr, serializeFn) {
        const bufPtrPtr = malloc(ptrSize);
        const lenPtr    = malloc(ptrSize);
        serializeFn(ptr, bufPtrPtr, lenPtr);
        const bufPtr = dv.getUint32(bufPtrPtr, true);
        const len    = dv.getUint32(lenPtr, true);
        const bytes  = new Uint8Array(memory.buffer, bufPtr, len).slice();
        go_buf_free(bufPtr);
        free(bufPtrPtr);
        free(lenPtr);
        return btoa(String.fromCharCode(...bytes));
      }

      return {
        publicKey:  serialize(pkPtr, instance.exports.go_pk_serialize),
        privateKey: serialize(skPtr, instance.exports.go_sk_serialize)
      };
    }
  };
}