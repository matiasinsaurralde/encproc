// app.js

const baseUrl = 'http://localhost:1234'; // Set to '' for relative, or e.g. 'http://localhost:1234' for absolute URLs

let goEnc = null; // The object returned by initGoEncrypt

window.addEventListener('DOMContentLoaded', async () => {
  const status = document.getElementById("status");
  const qContainer = document.getElementById("questions");
  const form = document.getElementById("survey-form");
  const participationLinkDiv = document.getElementById("participation-link");
  const participationLinkInput = document.getElementById("participation-link-input");
  const copyLinkBtn = document.getElementById("copy-link-btn");

  // 1) Initialize Go WASM encryption
  try {
    status.textContent = 'Loading encryption library...';
    goEnc = await window.initGoEncrypt(`${baseUrl}/static/go-encrypt.wasm`);
    status.textContent = 'Encryption library loaded.';
    status.style.color = '#28a745';
  } catch (err) {
    console.error('WASM init error:', err);
    status.textContent = 'Failed to load encryption library.';
    status.style.color = '#dc3545';
    return;
  }

  // 2) Get stream ID from URL
  function getStreamIdFromPath() {
    const match = window.location.pathname.match(/^\/stream\/([^\/]+)\/contribute/);
    return match ? match[1] : null;
  }
  const streamId = getStreamIdFromPath();
  if (!streamId) {
    status.textContent = "Stream ID not found in URL!";
    status.style.color = '#dc3545';
    return;
  }

  // 3) Display participation link using baseUrl
  if (participationLinkDiv && participationLinkInput && copyLinkBtn) {
    const link = `${baseUrl}/stream/${streamId}/contribute`;
    participationLinkInput.value = link;
    participationLinkDiv.style.display = 'block';
    copyLinkBtn.addEventListener('click', () => {
      participationLinkInput.select();
      participationLinkInput.setSelectionRange(0, 99999); // For mobile
      document.execCommand('copy');
      copyLinkBtn.textContent = "Copied!";
      setTimeout(() => { copyLinkBtn.textContent = "Copy"; }, 1200);
    });
  }

  // 4) Fetch public key and aux (questions)
  let pubkey, questions;
  try {
    status.textContent = "Loading survey details...";
    const raw = await fetch(`${baseUrl}/public-key/${streamId}`).then(r => r.json());
    pubkey = raw.publicKey;
    questions = raw.aux;
    if (!questions) throw new Error("No aux data found for this stream.");
    status.textContent = "";
  } catch (err) {
    status.textContent = "Failed to load survey details: " + err;
    status.style.color = '#dc3545';
    return;
  }

  // 5) Render number questions with min/max boundaries
  questions.forEach((q, i) => {
    const card = document.createElement("div");
    card.className = "question-card";

    if (q.type === "number") {
      card.innerHTML = `
        <label for="q${i}">${q.question}</label>
        <input type="number" id="q${i}" min="${q.min}" max="${q.max}" required />
        <div class="minmax-info">Allowed range: <strong>${q.min}</strong> to <strong>${q.max}</strong></div>
      `;
    } else {
      // fallback for any other type (should not happen)
      card.innerHTML = `
        <label for="q${i}">${q.question}</label>
        <input type="text" id="q${i}" required />
      `;
    }

    qContainer.appendChild(card);
  });

  // 6) On submit: check min/max, push answers, encrypt, then POST
  form.addEventListener("submit", async e => {
    e.preventDefault();
    status.textContent = "Encrypting…";
    status.style.color = "";

    try {
      let valid = true;
      questions.forEach((q, i) => {
        const input = document.getElementById(`q${i}`);
        let rawVal = input.value;
        const num = Number(rawVal);

        // Check min/max boundaries
        if (typeof q.min === "number" && num < q.min) {
          input.style.borderColor = '#dc3545';
          valid = false;
        } else if (typeof q.max === "number" && num > q.max) {
          input.style.borderColor = '#dc3545';
          valid = false;
        } else {
          input.style.borderColor = '';
        }

        const err = goEnc.push(num);
        if (typeof err === "string" && err && !/success/i.test(err)) {
          throw new Error(err);
        }
      });

      if (!valid) {
        status.textContent = "Please enter values within the allowed ranges.";
        status.style.color = '#dc3545';
        return;
      }

      const ctB64 = goEnc.encrypt(pubkey);
      if (ctB64.startsWith("Error")) throw new Error(ctB64);

      const resp = await fetch(`${baseUrl}/contribute/aggregate`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ ct: ctB64, id: streamId })
      });

      if (resp.ok) {
        status.textContent = "Thank you! Your encrypted response has been sent.";
        status.style.color = '#28a745';
        form.reset();
      } else {
        status.textContent = "Submission failed – please try again.";
        status.style.color = '#dc3545';
      }
    } catch (err) {
      status.textContent = "Encryption or submission failed: " + err;
      status.style.color = '#dc3545';
    }
  });
});
