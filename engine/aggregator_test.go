//go:build !openfhe

package main

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/bgv"
)

func TestInitAggregator_Success(t *testing.T) {
	params := SetupTestParamsLATTIGO()
	pk := rlwe.NewPublicKey(params)
	pkBytes, err := pk.MarshalBinary()
	assert.NoError(t, err)
	paramsLit := params.ParametersLiteral()
	paramsJSON, err := json.Marshal(paramsLit)
	calc := DummyCalculator()
	aggr, err := calc.initAggregator(pkBytes, string(paramsJSON))
	assert.NoError(t, err)
	assert.NotNil(t, aggr)
	assert.NotNil(t, aggr.pk)
	assert.NotNil(t, aggr.params)
	assert.NotNil(t, aggr.eval)
}

func TestInitAggregator_BadParamsJSON(t *testing.T) {
	calc := DummyCalculator()
	_, err := calc.initAggregator([]byte{}, "{bad json")
	assert.Error(t, err)
}

func TestInitAggregator_BadPK(t *testing.T) {
	params := SetupTestParamsLATTIGO()
	paramsLit := params.ParametersLiteral()
	paramsJSON, err := json.Marshal(paramsLit)
	assert.NoError(t, err)

	calc := DummyCalculator()
	_, err = calc.initAggregator([]byte("badpk"), string(paramsJSON))
	assert.Error(t, err)
}

func TestSnapshotAggregate_NilCiphertext(t *testing.T) {
	agg := &aggregator{
		logger: DummyCalculator().logger,
	}
	result, err := agg.snapshotAggregate()
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestAggregate_FirstCiphertext(t *testing.T) {
	params := SetupTestParamsLATTIGO()
	ct := rlwe.NewCiphertext(params, 1, params.MaxLevel())
	ctBytes, err := ct.MarshalBinary()
	assert.NoError(t, err)

	agg := &aggregator{
		logger: DummyCalculator().logger,
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
		logger: DummyCalculator().logger,
	}
	err := agg.aggregate([]byte("badct"))
	assert.Error(t, err)
}

func TestAggregate_AddCiphertext(t *testing.T) {
	params := SetupTestParamsLATTIGO()
	ct1 := rlwe.NewCiphertext(params, 1, params.MaxLevel())
	ct2 := rlwe.NewCiphertext(params, 1, params.MaxLevel())
	ct1Bytes, err := ct1.MarshalBinary()
	assert.NoError(t, err)
	ct2Bytes, err := ct2.MarshalBinary()
	assert.NoError(t, err)

	agg := &aggregator{
		logger: DummyCalculator().logger,
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
