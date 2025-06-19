package main

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"

	"github.com/collapsinghierarchy/encproc/models/mocks"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/bgv"
)

const (
	testVal1 = uint64(7)
	testVal2 = uint64(5)
)

func SetupTestParamsLATTIGO() bgv.Parameters {
	params, err := bgv.NewParametersFromLiteral(
		bgv.ParametersLiteral{
			LogN:             12,
			LogQ:             []int{58},
			PlaintextModulus: 0x10001,
		})
	if err != nil {
		panic(err)
	}
	return params
}

// --- Test helper: setup server ---
func setupTestServerLATTIGO() (*calculator, bgv.Parameters, *rlwe.PublicKey, *rlwe.SecretKey, func()) {
	params := SetupTestParamsLATTIGO()
	pk := rlwe.NewPublicKey(params)
	sk := rlwe.NewSecretKey(params)
	calc := NewCalc()

	mux := http.NewServeMux()
	mux.HandleFunc("/create-stream", calc.createStream)
	mux.HandleFunc("/contribute/aggregate", calc.contributeAggregate)
	mux.HandleFunc("/snapshot/aggregate/{id}", calc.returnAggregate)
	srv := httptest.NewServer(mux)

	teardown := func() {
		srv.Close()
	}
	return calc, params, pk, sk, teardown
}

// dummyCalculator returns a calculator with a basic logger.
func DummyCalculator() *calculator {
	l := slog.New(slog.NewTextHandler(io.Discard, nil)) // Use io.Discard for silent logging in tests
	return &calculator{logger: l}
}

func silentLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func NewCalc() *calculator {
	return &calculator{
		logger:     silentLogger(),
		calc_model: &mocks.EncProcModel{},
	}
}
