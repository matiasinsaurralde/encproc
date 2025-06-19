//go:build !openfhe

package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"errors"

	"github.com/stretchr/testify/require"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/bgv"
)

var ErrNotFound = errors.New("not found")

// --- Integration test for handlers with regular aggregator ---
func TestHandlers_Aggregator_AggregationFlow(t *testing.T) {
	calc, params, pk, sk, teardown := setupTestServerLATTIGO()
	defer teardown()

	// 1. Create stream
	pkBytes, err := pk.MarshalBinary()
	require.NoError(t, err)
	paramsJSON, err := json.Marshal(params.ParametersLiteral())
	require.NoError(t, err)
	reqBody, _ := json.Marshal(map[string]string{
		"pk":     base64.StdEncoding.EncodeToString(pkBytes),
		"params": string(paramsJSON),
	})
	req := httptest.NewRequest(http.MethodPost, "/create-stream", bytes.NewReader(reqBody))
	w := httptest.NewRecorder()
	calc.createStream(w, req)
	require.Equal(t, http.StatusOK, w.Result().StatusCode)

	var createResp struct{ ID string }
	require.NoError(t, json.NewDecoder(w.Body).Decode(&createResp))
	require.NotEmpty(t, createResp.ID)

	// After create-stream
	t.Logf("Stream ID: %s", createResp.ID)

	// 2. Contribute two ciphertexts
	encoder := bgv.NewEncoder(params)
	encryptor := bgv.NewEncryptor(params, pk)
	for _, v := range []uint64{testVal1, testVal2} {
		pt := bgv.NewPlaintext(params, params.MaxLevel())
		encoder.Encode([]uint64{v}, pt)
		ct, err := encryptor.EncryptNew(pt)
		require.NoError(t, err)
		ctBytes, err := ct.MarshalBinary()
		require.NoError(t, err)

		contribBody, _ := json.Marshal(map[string]string{
			"ct": base64.StdEncoding.EncodeToString(ctBytes),
			"id": createResp.ID,
		})
		req := httptest.NewRequest(http.MethodPost, "/contribute/aggregate", bytes.NewReader(contribBody))
		w := httptest.NewRecorder()
		calc.contributeAggregate(w, req)
		require.Equal(t, http.StatusOK, w.Result().StatusCode)
		// Debug: print agg_map keys
	}

	// Before each contribution and snapshot
	t.Logf("Using stream ID: %s", createResp.ID)

	// 3. Snapshot aggregate
	req = httptest.NewRequest(http.MethodGet, "/snapshot/aggregate/"+createResp.ID, nil)
	req.SetPathValue("id", createResp.ID)
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

	// 4. Decrypt & verify
	decryptor := bgv.NewDecryptor(params, sk)
	ctAggr := &rlwe.Ciphertext{}
	require.NoError(t, ctAggr.UnmarshalBinary(ctAggBytes))
	ptRes := bgv.NewPlaintext(params, params.MaxLevel())
	decryptor.Decrypt(ctAggr, ptRes)
	values := make([]uint64, 1)
	err = encoder.Decode(ptRes, values)
	require.NoError(t, err)
	require.Equal(t, testVal1+testVal2, values[0])
}
