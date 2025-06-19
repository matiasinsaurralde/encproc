package he

import "github.com/tuneinsight/lattigo/v6/schemes/bgv"

func SetupParams() bgv.Parameters {
	var err error
	var params bgv.Parameters
	if params, err = bgv.NewParametersFromLiteral(
		bgv.ParametersLiteral{
			LogN:             12,        // log2(ring degree)
			LogQ:             []int{58}, // log2(primes Q) (ciphertext modulus)
			PlaintextModulus: 0x10001,   // log2(scale)
		}); err != nil {
		panic(err)
	}
	return params
}

// SetupParams returns the BGV parameters used in the HE scheme.
