// go_genkey.js
// Initializes Go-WASM key generation and exposes generateKeypair()
import { Go } from './wasm_exec.js';

export async function initGoGenKey(wasmUrl) {
  const go = new Go();
  const resp = await fetch(wasmUrl);
  const { instance } = await WebAssembly.instantiateStreaming(resp, go.importObject);
  go.run(instance);

  // Go exports exportKeypair as the keygen function
  return {
    generateKeypair: window.exportKeypair
  };
}