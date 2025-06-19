//go:build openfhe

package main

import (
	"net/http"
	"net/http/httptest"

	ofhe "github.com/collapsinghierarchy/openfhe-go"
)

const (
	testDepth = uint32(1)     // matches initAggregator’s hard-coded depth
	testT     = uint64(65537) // plaintext modulus
)

func setupTestServerOFHE(testDepth uint32, testT uint64) (*calculator, *ofhe.Context, *ofhe.PublicKey, *ofhe.SecretKey, func()) {
	// 1) crypto context + keys
	ctx := ofhe.NewBGVRNS(testDepth, testT)
	pk, sk, err := ctx.KeyGenPtr()
	if err != nil {
		panic(err)
	}

	calc := NewCalc()

	// 3) routes & httptest server
	mux := http.NewServeMux()
	mux.HandleFunc("/create-stream", calc.createStream)
	mux.HandleFunc("/contribute/aggregate", calc.contributeAggregate)
	mux.HandleFunc("/snapshot/aggregate/{id}", calc.returnAggregate)

	srv := httptest.NewServer(mux)

	// 4) cleanup closure
	teardown := func() {
		srv.Close()
		ctx.Free()
		pk.Free()
		sk.Free()
	}

	return calc, ctx, pk, sk, teardown
}
