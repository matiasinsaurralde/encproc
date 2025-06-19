//go:build openfhe

package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	ofhe "github.com/collapsinghierarchy/openfhe-go"
	"github.com/stretchr/testify/require"
)

// -----------------------------------------------------------------------------
// Full happy‑path integration test: create stream → contribute → snapshot.
// -----------------------------------------------------------------------------

func TestHandlers_OpenFHE_AggregationFlow(t *testing.T) {
	calc, ctx, pk, sk, teardown := setupTestServerOFHE(testDepth, testT)
	defer teardown()

	// --- 1. create stream -----------------------------------------------------
	pkBytes, err := pk.Serialize()
	require.NoError(t, err)

	reqBody, _ := json.Marshal(map[string]string{
		"pk": base64.StdEncoding.EncodeToString(pkBytes),
	})
	req := httptest.NewRequest(http.MethodPost, "/create-stream", bytes.NewReader(reqBody))
	w := httptest.NewRecorder()
	calc.createStream(w, req)
	require.Equal(t, http.StatusOK, w.Result().StatusCode)

	var createResp struct{ ID string }
	require.NoError(t, json.NewDecoder(w.Body).Decode(&createResp))
	require.NotEmpty(t, createResp.ID)
	t.Logf("Stream ID: %s", createResp.ID)

	// --- 2. contribute two ciphertexts ---------------------------------------

	for _, plaintext := range []uint64{testVal1, testVal2} {
		ct, err := ctx.EncryptU64ToPtr(pk, plaintext)
		require.NoError(t, err)
		ctBytes, err := ct.Serialize()
		require.NoError(t, err)

		contribBody, _ := json.Marshal(map[string]string{
			"ct": base64.StdEncoding.EncodeToString(ctBytes),
			"id": createResp.ID, // Pass ID in JSON body
		})
		req := httptest.NewRequest(http.MethodPost, "/contribute/aggregate", bytes.NewReader(contribBody))
		w := httptest.NewRecorder()
		calc.contributeAggregate(w, req)
		require.Equal(t, http.StatusOK, w.Result().StatusCode)
	}
	t.Logf("Using stream ID: %s", createResp.ID)

	// --- 3. snapshot aggregate ----------------------------------------------

	req = httptest.NewRequest(http.MethodGet, "/snapshot/aggregate/"+createResp.ID, nil)
	req.SetPathValue("id", createResp.ID) // Set path value for handler
	w = httptest.NewRecorder()
	calc.returnAggregate(w, req)
	require.Equal(t, http.StatusOK, w.Result().StatusCode)

	var snapshotResp struct {
		CtAggBase64 string `json:"ct_aggr_byte_base64"`
	}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&snapshotResp))
	require.NotEmpty(t, snapshotResp.CtAggBase64)

	ctAggBytes, err := base64.StdEncoding.DecodeString(snapshotResp.CtAggBase64)
	require.NoError(t, err)

	// --- 4. decrypt & verify --------------------------------------------------

	ctAgg := ofhe.DeserializeCiphertext(ctx, ctAggBytes)
	sum, err := ctx.DecryptU64FromPtr(sk, ctAgg)
	require.NoError(t, err)
	require.Equal(t, testVal1+testVal2, sum)
}

// Compile‑time assertion that we really implement the full contract.
// compile‑time check temporarily disabled because EncProcModel is declared as *EncProcModel in production code
// once EncProcModel is an interface value you can restore: var _ EncProcModel = (*InMemoryModel)(nil)
