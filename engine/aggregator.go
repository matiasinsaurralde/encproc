//go:build !openfhe

package main

import (
	"encoding/json"
	"errors"
	"log/slog"
	"sync"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/bgv"
)

type aggregator struct {
	logger  *slog.Logger
	pk      *rlwe.PublicKey
	params  bgv.Parameters
	eval    *bgv.Evaluator
	ctr     int
	ct_aggr *rlwe.Ciphertext
	mu      sync.Mutex
}

func (agg *aggregator) snapshotAggregate() ([]byte, error) {
	agg.mu.Lock()
	defer agg.mu.Unlock()
	if agg.ct_aggr == nil {
		if agg.logger != nil {
			agg.logger.Error("ct_aggr is nil, cannot serialize")
		}
		return nil, errors.New("no aggregate available")
	}
	// Serialize the ciphertext
	ct_aggr_byte, err := agg.ct_aggr.MarshalBinary()
	if err != nil {
		if agg.logger != nil {
			agg.logger.Error("failed to serialize ciphertext")
		}
		return nil, err
	}
	return ct_aggr_byte, nil
}

func (calc *calculator) initAggregator(pk []byte, params string) (*aggregator, error) {
	aggr := aggregator{ctr: 0, logger: calc.logger}
	params_lit := bgv.ParametersLiteral{}
	// Deserialize the JSON into the struct
	err := json.Unmarshal([]byte(params), &params_lit)
	if err != nil {
		if calc.logger != nil {
			calc.logger.Error("Error unmarshaling JSON Literal Params")
		}
		return &aggr, err
	}
	aggr.params, err = bgv.NewParametersFromLiteral(params_lit)
	if err != nil {
		if calc.logger != nil {
			calc.logger.Error("Error Converting Param Literatl to BGV.params")
		}
		return &aggr, err
	}
	aggr.pk = rlwe.NewPublicKey(aggr.params)
	err = aggr.pk.UnmarshalBinary(pk)
	if err != nil {
		if calc.logger != nil {
			calc.logger.Error("Error Converting pk []byte to rlwe.PublicKey")
		}
		return &aggr, err
	}
	aggr.eval = bgv.NewEvaluator(aggr.params, nil)

	return &aggr, nil
}

func (agg *aggregator) aggregate(ct_bin []byte) error {
	agg.mu.Lock()
	defer agg.mu.Unlock()

	var err error
	ct := &rlwe.Ciphertext{}
	err = ct.UnmarshalBinary(ct_bin)
	if err != nil {
		agg.logger.Error("failed to deserialize ciphertext")
		return err
	}

	if agg.ct_aggr == nil {
		agg.ct_aggr = ct
		agg.ctr++
		return nil
	}

	if err := agg.eval.Add(agg.ct_aggr, ct, agg.ct_aggr); err != nil {
		agg.logger.Error("failed to add ciphertext", "error", err)
		return err
	}

	agg.ctr++
	return nil
}
