// ofh_encrypt.js
// Initializes OpenFHE-WASM encryption and exposes encrypt()
export async function initOFHEncrypt(wasmUrl, depth = 1, t = 0x10001) {
  const resp = await fetch(wasmUrl);
  const { instance } = await WebAssembly.instantiateStreaming(resp, {});
  const { memory, malloc, free, go_bgvrns_new, go_pk_deserialize, go_encrypt_u64_ptr_out, go_ct_ser, go_buf_free, go_ct_free } = instance.exports;
  const dv = new DataView(memory.buffer);
  const ptrSize = 4;

  const ctx = go_bgvrns_new(depth, t);

  return {
    encrypt: (publicKeyB64, value) => {
      const pkBytes = Uint8Array.from(atob(publicKeyB64), c => c.charCodeAt(0));
      const buf     = malloc(pkBytes.length);
      new Uint8Array(memory.buffer, buf, pkBytes.length).set(pkBytes);
      const pkPtr   = go_pk_deserialize(ctx, buf, pkBytes.length);
      free(buf);

      const ctPtr   = go_encrypt_u64_ptr_out(ctx, pkPtr, BigInt(value));
      // serialize ciphertext
      const bufPtrPtr = malloc(ptrSize);
      const lenPtr    = malloc(ptrSize);
      go_ct_ser(ctPtr, bufPtrPtr, lenPtr);
      const bufPtr = dv.getUint32(bufPtrPtr, true);
      const len    = dv.getUint32(lenPtr, true);
      const bytes  = new Uint8Array(memory.buffer, bufPtr, len).slice();
      go_buf_free(bufPtr);
      go_ct_free(ctPtr);
      free(bufPtrPtr);
      free(lenPtr);
      return btoa(String.fromCharCode(...bytes));
    }
  };
}
