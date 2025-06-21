// Robust Go WASM loader for decryption

async function initGoDecrypt(wasmUrl) {
  const go = new Go();
  console.log('[WASM] fetching →', wasmUrl);

  const resp = await fetch(wasmUrl, { cache: 'no-cache' });
  if (!resp.ok) {
    throw new Error(`WASM fetch failed: ${resp.status} ${resp.statusText}`);
  }

  let instance;
  try {
    instance = await WebAssembly.instantiateStreaming(resp.clone(), go.importObject);
    console.log('[WASM] instantiateStreaming succeeded');
  } catch (streamErr) {
    console.warn('[WASM] instantiateStreaming failed, falling back to arrayBuffer', streamErr);
    const bytes = await resp.arrayBuffer();
    instance = await WebAssembly.instantiate(bytes, go.importObject);
    console.log('[WASM] instantiate(arrayBuffer) succeeded');
  }

  go.run(instance.instance);

  if (typeof window.eng_decrypt !== 'function') {
    throw new Error('eng_decrypt() not found on window');
  }

  return {
    decrypt: window.eng_decrypt
  };
}

window.initGoDecrypt = initGoDecrypt;