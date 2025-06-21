package main

import (
	"encoding/base64"
	"fmt"
	"strings"
	"syscall/js"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"

	gojshe "github.com/collapsinghierarchy/encproc/clientgojs/he"
)

func main() {
	he := &gojshe.HE{}
	he.Params = gojshe.SetupParams()

	// Exported encrypt function
	decryptFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// Validate input
		if len(args) < 2 {
			return js.ValueOf("Error: Ciphertext and Secret key are required")
		}

		// Get the input as a byte array
		ctBase64 := args[0].String()
		skBase64 := args[1].String()
		sk_bin, err := base64.StdEncoding.DecodeString(skBase64)
		if err != nil {
			return js.ValueOf("Error decoding Base64 Secret Key: " + err.Error())
		}
		ct_bin, err := base64.StdEncoding.DecodeString(ctBase64)
		if err != nil {
			return js.ValueOf("Error decoding Base64 Ciphertext: " + err.Error())
		}

		ct := &rlwe.Ciphertext{}
		err = ct.UnmarshalBinary(ct_bin)
		if err != nil {
			panic(err)
		}

		//unmarshall secret key
		sk := rlwe.NewSecretKey(he.Params)
		err = sk.UnmarshalBinary(sk_bin)
		if err != nil {
			panic(err)
		}

		// Perform decryption
		values, err := he.Decrypt_result(sk, ct)
		if err != nil {
			return js.ValueOf("Error during decryption: " + err.Error())
		}
		strValues := make([]string, len(values))
		for i, v := range values {
			strValues[i] = fmt.Sprintf("%d", v)
		}
		return strings.Join(strValues, ",")
	})

	// Expose the `encrypt` function to JavaScript
	js.Global().Set("eng_decrypt", decryptFunc)

	// Keep the WASM module running
	select {}
}
