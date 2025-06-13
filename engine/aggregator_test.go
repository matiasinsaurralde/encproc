package main

import (
	"encoding/json"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/bgv"
)

// setupTestParams provides deterministic test parameters for BGV.
func setupTestParams() bgv.Parameters {
	var err error
	var params bgv.Parameters
	if params, err = bgv.NewParametersFromLiteral(
		bgv.ParametersLiteral{
			LogN:             12,
			LogQ:             []int{58},
			PlaintextModulus: 0x10001,
		}); err != nil {
		panic(err)
	}
	return params
}

// dummyCalculator returns a calculator with a basic logger.
func dummyCalculator() *calculator {
	l := slog.New(slog.NewTextHandler(io.Discard, nil)) // Use io.Discard for silent logging in tests
	return &calculator{logger: l}
}

func TestInitAggregator_Success(t *testing.T) {
	params := setupTestParams()

	pk := rlwe.NewPublicKey(params)
	pkBytes, err := pk.MarshalBinary()
	assert.NoError(t, err)
	paramsLit := params.ParametersLiteral()
	paramsJSON, err := json.Marshal(paramsLit)
	calc := dummyCalculator()
	aggr, err := calc.initAggregator(pkBytes, string(paramsJSON))
	assert.NoError(t, err)
	assert.NotNil(t, aggr)
	assert.NotNil(t, aggr.pk)
	assert.NotNil(t, aggr.params)
	assert.NotNil(t, aggr.eval)
}

func TestInitAggregator_BadParamsJSON(t *testing.T) {
	calc := dummyCalculator()
	_, err := calc.initAggregator([]byte{}, "{bad json")
	assert.Error(t, err)
}

func TestInitAggregator_BadPK(t *testing.T) {
	params := setupTestParams()
	paramsLit := params.ParametersLiteral()
	paramsJSON, err := json.Marshal(paramsLit)
	assert.NoError(t, err)

	calc := dummyCalculator()
	_, err = calc.initAggregator([]byte("badpk"), string(paramsJSON))
	assert.Error(t, err)
}

func TestSnapshotAggregate_NilCiphertext(t *testing.T) {
	agg := &aggregator{
		logger: dummyCalculator().logger,
	}
	result := agg.snapshotAggregate()
	assert.Nil(t, result)
}

func TestAggregate_FirstCiphertext(t *testing.T) {
	params := setupTestParams()
	ct := rlwe.NewCiphertext(params, 1, params.MaxLevel())
	ctBytes, err := ct.MarshalBinary()
	assert.NoError(t, err)

	agg := &aggregator{
		logger: dummyCalculator().logger,
		params: params,
		eval:   bgv.NewEvaluator(params, nil),
	}
	err = agg.aggregate(ctBytes)
	assert.NoError(t, err)
	assert.NotNil(t, agg.ct_aggr)
	assert.Equal(t, 1, agg.ctr)
}

func TestAggregate_BadCiphertext(t *testing.T) {
	agg := &aggregator{
		logger: dummyCalculator().logger,
	}
	err := agg.aggregate([]byte("badct"))
	assert.Error(t, err)
}

func TestAggregate_AddCiphertext(t *testing.T) {
	params := setupTestParams()
	ct1 := rlwe.NewCiphertext(params, 1, params.MaxLevel())
	ct2 := rlwe.NewCiphertext(params, 1, params.MaxLevel())
	ct1Bytes, err := ct1.MarshalBinary()
	assert.NoError(t, err)
	ct2Bytes, err := ct2.MarshalBinary()
	assert.NoError(t, err)

	agg := &aggregator{
		logger: dummyCalculator().logger,
		params: params,
		eval:   bgv.NewEvaluator(params, nil),
	}
	// First aggregate
	err = agg.aggregate(ct1Bytes)
	assert.NoError(t, err)
	// Second aggregate (should trigger Add)
	err = agg.aggregate(ct2Bytes)
	assert.NoError(t, err)
	assert.Equal(t, 2, agg.ctr)
}
