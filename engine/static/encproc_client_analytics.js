/* analytics.example.com/encproc-tot.js  */
(function () {
  const tag      = document.currentScript;
  const streamId = tag.getAttribute('data-stream');
  if (!streamId) { console.error('[encproc] missing data-stream'); return; }

  const ORIGIN   = new URL(tag.src).origin;              
  const paths    = {
    wasmExec : `${ORIGIN}/wasm_exec.js`,
    goEnc    : `${ORIGIN}/go_encryption.js`,
    wasmBin  : `${ORIGIN}/encproc.wasm`,
    pubKey   : `${ORIGIN}/public-key/${streamId}`,
    ingest   : `${ORIGIN}/contribute/aggregate`
  };

  /* tiny helper to chain <script> loads */
  const load = src => new Promise((ok, err) => {
    const s = document.createElement('script');
    s.src = src; s.defer = true;
    s.onload = ok; s.onerror = () => err(new Error(`${src} failed`));
    document.head.appendChild(s);
  });

  (async () => {
    /* 1️⃣  guarantee wasm_exec.js + Go() */
    if (!window.Go)           await load(paths.wasmExec);
    /* 2️⃣  guarantee go_encryption.js + initGoEncrypt() */
    if (!window.initGoEncrypt) await load(paths.goEnc);

    /* 3️⃣  fetch public key and start WASM */
    const [{ publicKey }, enc] = await Promise.all([
      fetch(paths.pubKey).then(r => r.json()),
      window.initGoEncrypt(paths.wasmBin)
    ]);

    /* 4️⃣  start time-on-task counter */
    let start = performance.now(), sent = false;
    const flush = () => {
      if (sent) return; sent = true;
      const sec = ((performance.now() - start) / 1000) | 0;
      enc.push(sec);
      const ct = enc.encrypt(publicKey);
      const body = JSON.stringify({ id: streamId, ct });

      /* sendBeacon if possible; otherwise POST */
      if (!navigator.sendBeacon?.(paths.ingest, body)) {
        fetch(paths.ingest, { method: 'POST',
                              headers: { 'Content-Type': 'application/json' },
                              body });
      }
    };

    addEventListener('pagehide', flush, { once: true });
    addEventListener('visibilitychange', () =>
      document.visibilityState === 'hidden' && flush(), { once: true });
    addEventListener('beforeunload', flush, { once: true });

    /* expose manual flush for SPA route changes */
    window.encprocFlush = flush;
  })().catch(err => console.error('[encproc] loader error:', err));
})();
