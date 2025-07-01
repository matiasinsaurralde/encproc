const baseUrl = ''; // Set to '' for relative, or e.g. 'http://localhost:1234' for absolute URLs

// Initialize variables for keypair and stream ID
let keypair = null;
let latestStreamId = null;
let goKey = null;  // the object returned by initGoGenKey

window.addEventListener('DOMContentLoaded', async () => {
  const questionsList = document.getElementById('questions-list');
  const addBtn = document.getElementById('add-question-btn');
  const form = document.getElementById('survey-form');
  const keyStatus = document.getElementById('keypair-status');
  const result = document.getElementById('result');
  const downloadBtn = document.getElementById('download-key-btn');
  downloadBtn.style.display = 'none'; // Ensure hidden on load

  const participationLinkDiv = document.getElementById('participation-link');
  const participationLinkInput = document.getElementById('participation-link-input');
  const copyLinkBtn = document.getElementById('copy-link-btn');
  participationLinkDiv.style.display = 'none';

  // Only allow number questions with min/max
  function addQuestionField(val = '', minVal = '', maxVal = '') {
    const div = document.createElement('div');
    div.className = 'question-field';

    const input = document.createElement('input');
    input.type = 'text';
    input.className = 'question-input';
    input.placeholder = 'Enter your question...';
    input.value = val;
    input.required = true;

    const min = document.createElement('input');
    min.type = 'number';
    min.className = 'min-input';
    min.placeholder = 'Min';
    min.value = minVal;
    min.required = true;
    min.style.display = 'inline-block';

    const max = document.createElement('input');
    max.type = 'number';
    max.className = 'max-input';
    max.placeholder = 'Max';
    max.value = maxVal;
    max.required = true;
    max.style.display = 'inline-block';

    const removeBtn = document.createElement('button');
    removeBtn.type = 'button';
    removeBtn.className = 'remove-btn';
    removeBtn.innerHTML = '&times;';
    removeBtn.title = 'Remove this question';
    removeBtn.onclick = () => div.remove();

    div.append(input, min, max, removeBtn);
    questionsList.appendChild(div);
  }

  addQuestionField();
  addBtn.addEventListener('click', () => addQuestionField());

  // --- Generate keypair via WASM ---
  async function generateKeypair() {
    keyStatus.textContent = 'Generating encryption keypair...';
    keyStatus.style.color = '#888';
    try {
      const kp = await goKey.generateKeypair();
      keypair = kp;
      keyStatus.textContent = 'Keypair generated!';
      keyStatus.style.color = '#28a745';
      return kp;
    } catch (e) {
      console.error('Keypair error:', e);
      keyStatus.textContent = 'Failed to generate keypair';
      keyStatus.style.color = '#dc3545';
      throw e;
    }
  }

  // --- Form submission ---
  form.addEventListener('submit', async (e) => {
    e.preventDefault();

    const nodes = Array.from(document.querySelectorAll('.question-field'));
    const questions = [];
    let valid = true;

    nodes.forEach(div => {
      const qInput = div.querySelector('.question-input');
      const qText = qInput.value.trim();
      const minVal = Number(div.querySelector('.min-input').value);
      const maxVal = Number(div.querySelector('.max-input').value);

      if (!qText) {
        valid = false;
        qInput.style.borderColor = '#dc3545';
      } else {
        qInput.style.borderColor = '';
      }

      if (isNaN(minVal) || isNaN(maxVal) || minVal >= maxVal) {
        valid = false;
        div.querySelector('.min-input').style.borderColor = '#dc3545';
        div.querySelector('.max-input').style.borderColor = '#dc3545';
      } else {
        div.querySelector('.min-input').style.borderColor = '';
        div.querySelector('.max-input').style.borderColor = '';
        questions.push({ question: qText, type: 'number', min: minVal, max: maxVal });
      }
    });

    if (!valid || questions.length === 0) {
      alert('Please correct all fields (each question needs text and a valid number range).');
      return;
    }

    if (!keypair) {
      try {
        await generateKeypair();
      } catch {
        return;
      }
    }

    keyStatus.textContent = 'Creating encrypted survey...';
    keyStatus.style.color = '#888';

    try {
      const resp = await fetch(`${baseUrl}/create-stream`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ pk: keypair.publicKey, "aux" : questions })
      });
      if (!resp.ok) throw new Error(resp.status);
      const data = await resp.json();
      latestStreamId = data.id;

      // Show participation link at the top, copyable
      const link = `${baseUrl}/stream/${data.id}/contribute`;
      participationLinkInput.value = link;
      participationLinkDiv.style.display = 'block';

      // Show download button
      downloadBtn.style.display = 'inline-block'; // Only after successful creation

      keyStatus.textContent = '';
      result.style.display = 'block';
      result.innerHTML = `<strong>Survey created! Don't forget to download the Access File. Without it the answers will be lost.</strong>`;

      // HIDE the form after successful creation
      form.style.display = 'none';

      // Optionally, keep the download button enabled
      downloadBtn.disabled = false;

    } catch (err) {
      keyStatus.textContent = 'Survey creation failed: ' + err;
      keyStatus.style.color = '#dc3545';
    }
  });

  // Copy link functionality
  copyLinkBtn.addEventListener('click', () => {
    participationLinkInput.select();
    participationLinkInput.setSelectionRange(0, 99999); // For mobile
    document.execCommand('copy');
    copyLinkBtn.textContent = "Copied!";
    setTimeout(() => { copyLinkBtn.textContent = "Copy"; }, 1200);
  });

  // --- Download key material ---
  downloadBtn.addEventListener('click', () => {
    if (!keypair || !latestStreamId) {
      alert('No keypair or stream ID available. Create a survey first.');
      return;
    }
    const material = {
      stream_id: latestStreamId,
      private_key: keypair.privateKey,
      public_key: keypair.publicKey
    };
    const blob = new Blob([JSON.stringify(material, null, 2)], { type: 'application/json' });
    const a = document.createElement('a');
    a.href = URL.createObjectURL(blob);
    a.download = `key_material_${latestStreamId}.json`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
  });

  // 1) Initialize Go WASM key generator
  try {
    keyStatus.textContent = 'Loading crypto library...';
    goKey = await window.initGoGenKey(`${baseUrl}/static/go-genkey.wasm`);
    keyStatus.textContent = 'Crypto library loaded.';
    keyStatus.style.color = '#28a745';
  } catch (err) {
    console.error('WASM init error:', err);
    keyStatus.textContent = 'Failed to load crypto library.';
    keyStatus.style.color = '#dc3545';
    return;
  }

  async function updateThumbsUpCount() {
    try {
      const resp = await fetch(`${baseUrl}/thumbs-up`);
      const data = await resp.json();
      document.getElementById('thumbs-up-count').textContent = data.count;
    } catch {}
  }

  document.getElementById('thumbs-up-btn').addEventListener('click', async () => {
    try {
      const resp = await fetch(`${baseUrl}/thumbs-up`, { method: 'POST' });
      const data = await resp.json();
      document.getElementById('thumbs-up-count').textContent = data.count;
    } catch {}
  });

  // Initialize on page load
  updateThumbsUpCount();
});
