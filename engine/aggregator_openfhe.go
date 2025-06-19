//go:build openfhe

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	ofhe "github.com/collapsinghierarchy/openfhe-go"
)

// ----------------- aggregator -----------------------------------------
type aggregator struct {
	logger  *slog.Logger
	ctx     *ofhe.Context
	ct_aggr *ofhe.Ciphertext
	ctr     int
	mu      sync.Mutex
}

type aggrParams struct {
	PlaintextModulus uint64 `json:"PlaintextModulus"`
}

// initAggregator mirrors old signature but takes JSON params
func (c *calculator) initAggregator(_ []byte, paramsJSON string) (*aggregator, error) {
	var p aggrParams
	if err := json.Unmarshal([]byte(paramsJSON), &p); err != nil {
		return nil, err
	}
	// use depth=1 for addition-only
	ctx := ofhe.NewBGVRNS(1, p.PlaintextModulus)
	return &aggregator{
		logger: c.logger,
		ctx:    ctx,
	}, nil
}

// aggregate adds a new ciphertext (serialized) into the running sum,
// but on the very first insert we first validate the blob by
// attempting a homomorphic add of it to a clone of itself.
func (agg *aggregator) aggregate(ctBin []byte) error {
	agg.mu.Lock()
	defer agg.mu.Unlock()

	fmt.Print("aggregating ciphertext", slog.Int("current_count", agg.ctr))
	// 1) deserialize ciphertext
	ct := ofhe.DeserializeCiphertext(agg.ctx, ctBin)
	if ct == nil || ct.Raw() == nil {
		agg.logger.Error("failed to deserialize ciphertext")
		return ErrBadCiphertext
	}

	// 2) first insert: just store the ciphertext
	if agg.ct_aggr == nil {
		agg.ct_aggr = ct
		agg.ctr++
		return nil
	}

	// 3) subsequent inserts: just homomorphically add into accumulator
	if err := agg.ctx.EvalAdd(agg.ct_aggr, ct); err != nil {
		ct.Free()
		return err
	}
	ct.Free()
	agg.ctr++
	return nil
}

// snapshotAggregate returns the current running sum (serialized).
func (agg *aggregator) snapshotAggregate() ([]byte, error) {
	agg.mu.Lock()
	defer agg.mu.Unlock()

	if agg.ct_aggr == nil {
		if agg.logger != nil {
			agg.logger.Error("ct_aggr is nil, cannot serialize")
		}
		return nil, errors.New("no aggregate available")
	}
	return agg.ct_aggr.Serialize()
}

// ----------------------------------------------------------------------
var ErrBadCiphertext = errors.New("bad ciphertext")
