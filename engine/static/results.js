const baseUrl = 'http://localhost:1234'; // Set to '' for relative, or e.g. 'http://localhost:1234' for absolute URLs

let goDec = null;

window.addEventListener('DOMContentLoaded', async () => {
  const status = document.getElementById('status');
  const fileInput = document.getElementById('access-file-input');
  const resultsContainer = document.getElementById('results-container');

  // 1) Initialize Go WASM decryption
  try {
    status.textContent = 'Loading decryption library...';
    goDec = await window.initGoDecrypt(`${baseUrl}/static/go-decrypt.wasm`);
    status.textContent = 'Decryption library loaded.';
    status.style.color = '#28a745';
  } catch (err) {
    console.error('WASM init error:', err);
    status.textContent = 'Failed to load decryption library.';
    status.style.color = '#dc3545';
    return;
  }

  // 2) Handle file input
  fileInput.addEventListener('change', async (e) => {
    const file = e.target.files[0];
    if (!file) return;

    status.textContent = 'Reading access file...';
    status.style.color = '';

    try {
      const text = await file.text();
      const accessData = JSON.parse(text);

      // Expecting: { "stream_id": "...", "private_key": "...", ... }
      const streamId = accessData.stream_id;
      const secretKey = accessData.private_key;
      if (!streamId || !secretKey) {
        throw new Error('Access file missing stream_id or private_key');
      }

      status.textContent = 'Fetching survey results...';

      // Fetch snapshot/aggregate/{streamId} to get ciphertexts and aux
      let snapshot;
      try {
        const resp = await fetch(`${baseUrl}/snapshot/aggregate/${streamId}`);
        if (!resp.ok) throw new Error(`HTTP ${resp.status}`);
        snapshot = await resp.json();
      } catch (err) {
        throw new Error('Failed to fetch snapshot: ' + err);
      }

      // Expecting: { ct_aggr_byte_base64: "...", aux: [...], sample_size: N }
      const ctAgg = snapshot.ct_aggr_byte_base64;
      const questions = snapshot.aux || [];
      const sampleSize = snapshot.sample_size || 0;

      if (!ctAgg) {
        throw new Error('No aggregate ciphertext found in snapshot.');
      }
      if (!Array.isArray(questions) || questions.length === 0) {
        throw new Error('No questions (aux) found in snapshot.');
      }
      if (!sampleSize || sampleSize <= 0) {
        throw new Error('Sample size is missing or zero.');
      }

      resultsContainer.innerHTML = '';
      let result;
      try {
        result = goDec.decrypt(ctAgg, secretKey);
      } catch (err) {
        result = 'Decryption error';
      }

      // Split the result string into an array of values
      const values = typeof result === "string" ? result.split(",") : [];

      // Compute and display the mean for each question
      questions.forEach((q, i) => {
        const card = document.createElement('div');
        card.className = 'result-card';
        const value = values[i] !== undefined ? parseFloat(values[i]) : NaN;
        let mean = "(no data)";
        if (!isNaN(value) && sampleSize > 0) {
          mean = (value / sampleSize).toFixed(2);
        }
        card.innerHTML = `<h3>${q.question || `Question ${i+1}`}</h3><div><strong>Mean: ${mean}</strong></div>`;
        resultsContainer.appendChild(card);
      });

      status.textContent = `Results decrypted. (Sample size: ${sampleSize})`;
      status.style.color = '#28a745';
    } catch (err) {
      status.textContent = 'Failed to read or decrypt: ' + err;
      status.style.color = '#dc3545';
      resultsContainer.innerHTML = '';
    }
  });
});