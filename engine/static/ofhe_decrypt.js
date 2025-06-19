// ofh_decrypt.js
// Initializes OpenFHE-WASM decryption and exposes decrypt()
export async function initOFHDecrypt(wasmUrl, depth = 1, t = 0x10001) {
  const resp = await fetch(wasmUrl);
  const { instance } = await WebAssembly.instantiateStreaming(resp, {});
  const { memory, malloc, free, go_bgvrns_new, go_sk_deserialize, go_ct_deser, go_decrypt_u64_ptr, go_sk_free, go_ct_free } = instance.exports;
  const dv = new DataView(memory.buffer);
  const ptrSize = 4;

  const ctx = go_bgvrns_new(depth, t);

  return {
    decrypt: (secretKeyB64, ctB64) => {
      const skBytes = Uint8Array.from(atob(secretKeyB64), c => c.charCodeAt(0));
      let buf = malloc(skBytes.length);
      new Uint8Array(memory.buffer, buf, skBytes.length).set(skBytes);
      const skPtr = go_sk_deserialize(ctx, buf, skBytes.length);
      free(buf);

      const ctBytes = Uint8Array.from(atob(ctB64), c => c.charCodeAt(0));
      buf = malloc(ctBytes.length);
      new Uint8Array(memory.buffer, buf, ctBytes.length).set(ctBytes);
      const ctPtr = go_ct_deser(ctx, buf, ctBytes.length);
      free(buf);

      const val = go_decrypt_u64_ptr(ctx, skPtr, ctPtr);
      go_sk_free(skPtr);
      go_ct_free(ctPtr);
      return Number(val);
    }
  };
}
