// go_decrypt.js
// Initializes Go-WASM decryption and exposes decrypt()
import { Go } from './wasm_exec.js';

export async function initGoDecrypt(wasmUrl) {
  const go = new Go();
  const resp = await fetch(wasmUrl);
  const { instance } = await WebAssembly.instantiateStreaming(resp, go.importObject);
  go.run(instance);

  return {
    decrypt: window.decrypt_result
  };
}