package main

import (
	"encoding/base64"
	"syscall/js"

	gojshe "github.com/collapsinghierarchy/encproc/clientgojs/he"
)

type inputEncoder struct {
	input []uint64
}

func main() {
	heInstance := &gojshe.HE{}
	heInstance.Params = gojshe.SetupParams()

	inputEncoder := &inputEncoder{input: make([]uint64, 0, heInstance.Params.MaxSlots())}

	// Populate function exposed to JavaScript
	populateFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 1 {
			return js.ValueOf("Error: Input required")
		}

		// Überprüfe, ob der übergebene Wert vom Typ Number ist.
		if args[0].Type() != js.TypeNumber {
			return js.ValueOf("Error: Input must be a number")
		}

		value := args[0].Int()

		// Füge den Wert zum Slice hinzu.
		inputEncoder.input = append(inputEncoder.input, uint64(value))
		return js.ValueOf("Input updated successfully")
	})

	// Encrypt function exposed to JavaScript
	encryptFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 1 {
			return js.ValueOf("Error: Base64-encoded public key is required")
		}

		// Parse Base64-encoded public key
		pkBase64 := args[0].String()
		pkBin, err := base64.StdEncoding.DecodeString(pkBase64)
		if err != nil {
			return js.ValueOf("Error decoding Base64 public key: " + err.Error())
		}

		// Perform encryption
		ctBin, err := heInstance.EncryptInput(inputEncoder.input, pkBin)
		if err != nil {
			return js.ValueOf("Error during encryption: " + err.Error())
		}

		// Encode ciphertext to Base64
		ctBase64 := base64.StdEncoding.EncodeToString(ctBin)
		return js.ValueOf(ctBase64)
	})

	// Expose functions to JavaScript
	js.Global().Set("eng_push", populateFunc)
	js.Global().Set("eng_encrypt", encryptFunc)

	// Keep the WASM module running
	select {}
}
