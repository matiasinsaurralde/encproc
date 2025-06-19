//go:build openfhe

package main

import (
	"encoding/json"
	"testing"

	"github.com/collapsinghierarchy/openfhe-go"
	"github.com/stretchr/testify/assert"
)

func TestInitAggregator_Success(t *testing.T) {
	// JSON params
	params := map[string]uint64{"PlaintextModulus": testT}
	js, err := json.Marshal(params)
	assert.NoError(t, err)

	calc := NewCalc()
	agg, err := calc.initAggregator(nil, string(js))
	assert.NoError(t, err)
	assert.NotNil(t, agg)
	assert.NotNil(t, agg.ctx)
	assert.Zero(t, agg.ctr)
	assert.Nil(t, agg.ct_aggr)
}

func TestSnapshotAggregate_Empty(t *testing.T) {
	agg := &aggregator{logger: silentLogger()}
	out, err := agg.snapshotAggregate()
	assert.Error(t, err)
	assert.Nil(t, out)
}

func TestAggregate_FirstCiphertext(t *testing.T) {
	// client-side setup
	ctx := openfhe.NewBGVRNS(testDepth, testT)
	defer ctx.Free()
	pk, sk, err := ctx.KeyGenPtr()
	assert.NoError(t, err)
	defer pk.Free()
	defer sk.Free()

	// encrypt one value
	ct1, err := ctx.EncryptU64ToPtr(pk, testVal1)
	assert.NoError(t, err)
	defer ct1.Free()
	ct1Bytes, err := ct1.Serialize()
	assert.NoError(t, err)

	// aggregator setup
	params := map[string]uint64{"PlaintextModulus": testT}
	js, _ := json.Marshal(params)
	agg, err := NewCalc().initAggregator(nil, string(js))
	assert.NoError(t, err)

	// first aggregate
	err = agg.aggregate(ct1Bytes)
	assert.NoError(t, err)
	assert.NotNil(t, agg.ct_aggr)
	assert.Equal(t, 1, agg.ctr)
}

func TestAggregate_BadCiphertext(t *testing.T) {
	// aggregator with no prior state
	params := map[string]uint64{"PlaintextModulus": testT}
	js, _ := json.Marshal(params)
	agg, _ := NewCalc().initAggregator(nil, string(js))

	// feed it garbage
	err := agg.aggregate([]byte("not a ciphertext"))
	assert.Error(t, err)
}

func TestAggregate_AddCiphertexts(t *testing.T) {
	// client-side setup
	ctx := openfhe.NewBGVRNS(testDepth, testT)
	defer ctx.Free()
	pk, sk, err := ctx.KeyGenPtr()
	assert.NoError(t, err)
	defer pk.Free()
	defer sk.Free()

	// prepare two ciphertexts
	enc := func(v uint64) []byte {
		ct, e := ctx.EncryptU64ToPtr(pk, v)
		assert.NoError(t, e)
		defer ct.Free()
		b, e2 := ct.Serialize()
		assert.NoError(t, e2)
		return b
	}
	ct1 := enc(testVal1)
	ct2 := enc(testVal2)

	// aggregator setup
	params := map[string]uint64{"PlaintextModulus": testT}
	js, _ := json.Marshal(params)
	agg, _ := NewCalc().initAggregator(nil, string(js))

	// accumulate both
	assert.NoError(t, agg.aggregate(ct1))
	assert.NoError(t, agg.aggregate(ct2))
	assert.Equal(t, 2, agg.ctr)
}

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	ctx := openfhe.NewBGVRNS(testDepth, testT)
	defer ctx.Free()
	pk, sk, err := ctx.KeyGenPtr()
	assert.NoError(t, err)
	defer pk.Free()
	defer sk.Free()

	plaintext := uint64(12345)

	// Encrypt
	ct, err := ctx.EncryptU64ToPtr(pk, plaintext)
	assert.NoError(t, err)
	defer ct.Free()

	// Decrypt
	result, err := ctx.DecryptU64FromPtr(sk, ct)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, result)
}

func TestAggregateAndDecryptSnapshot(t *testing.T) {
	// client-side setup
	ctx := openfhe.NewBGVRNS(testDepth, testT)
	defer ctx.Free()
	pk, sk, err := ctx.KeyGenPtr()
	assert.NoError(t, err)
	defer pk.Free()
	defer sk.Free()

	// values to sum
	data := []uint64{testVal1, testVal2}
	var ctBins [][]byte
	for _, v := range data {
		ct, e := ctx.EncryptU64ToPtr(pk, v)
		assert.NoError(t, e)
		defer ct.Free()
		bin, e2 := ct.Serialize()
		assert.NoError(t, e2)
		ctBins = append(ctBins, bin)
	}

	// aggregator setup
	params := map[string]uint64{"PlaintextModulus": testT}
	js, _ := json.Marshal(params)
	agg, err := NewCalc().initAggregator(nil, string(js))
	assert.NoError(t, err)

	// aggregate them
	for i, bin := range ctBins {
		assert.NoError(t, agg.aggregate(bin))
		snapBytes, err := agg.snapshotAggregate()
		assert.NoError(t, err)
		snapCt := openfhe.DeserializeCiphertext(agg.ctx, snapBytes)
		result, err := ctx.DecryptU64FromPtr(sk, snapCt)
		t.Logf("After aggregation %d: decrypted sum = %d", i, result)
	}

	// snapshot and decrypt
	snapBytes, err := agg.snapshotAggregate()
	assert.NoError(t, err)
	assert.NotNil(t, snapBytes)

	snapCt := openfhe.DeserializeCiphertext(agg.ctx, snapBytes)
	assert.NotNil(t, snapCt)
	defer snapCt.Free()

	result, err := ctx.DecryptU64FromPtr(sk, snapCt)
	assert.NoError(t, err)
	assert.Equal(t, testVal1+testVal2, result)
}
