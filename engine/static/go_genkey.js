async function initGoGenKey(wasmUrl) {
  const go = new Go();
  console.log('[WASM] fetching →', wasmUrl);

  const resp = await fetch(wasmUrl, { cache: 'no-cache' });
  if (!resp.ok) {
    throw new Error(`WASM fetch failed: ${resp.status} ${resp.statusText}`);
  }

  // Try streaming first
  let instance;
  try {
    instance = await WebAssembly.instantiateStreaming(resp.clone(), go.importObject);
    console.log('[WASM] instantiateStreaming succeeded');
  } catch (streamErr) {
    console.warn('[WASM] instantiateStreaming failed, falling back to arrayBuffer', streamErr);

    // Clone again for arrayBuffer
    const clone = resp.clone();
    const bytes = await clone.arrayBuffer();

    // Debug: print magic number
    const magic = new Uint8Array(bytes.slice(0, 4));
    console.log('[WASM] first 4 bytes:', magic); 
    // Should log: Uint8Array [0, 97, 115, 109]

    instance = await WebAssembly.instantiate(bytes, go.importObject);
    console.log('[WASM] instantiate(arrayBuffer) succeeded');
  }

  go.run(instance.instance);
  console.log('[WASM] Go runtime started');

  if (typeof window.exportKeypair !== 'function') {
    throw new Error('exportKeypair() not found on window');
  }

  return { generateKeypair: window.exportKeypair };
}

window.initGoGenKey = initGoGenKey;
