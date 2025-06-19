// go_encryption.js
// Initializes Go-WASM encryption and exposes push() and encrypt()
import { Go } from './wasm_exec.js';

export async function initGoEncrypt(wasmUrl) {
  const go = new Go();
  const resp = await fetch(wasmUrl);
  const { instance } = await WebAssembly.instantiateStreaming(resp, go.importObject);
  go.run(instance);

  return {
    push:    window.eng_push,
    encrypt: window.eng_encrypt
  };
}