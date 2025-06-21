package main

import (
	"encoding/base64"
	"syscall/js"

	gojshe "github.com/collapsinghierarchy/encproc/clientgojs/he"
)

func main() {
	he := &gojshe.HE{}
	he.GenerateKeypair()

	export := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		pkBytes, skBytes, err := he.ExportBytes()
		if err != nil {
			return js.ValueOf("Error exporting bytes")
		}

		// Encode each key as base64 separately
		return map[string]interface{}{
			"publicKey":  base64.StdEncoding.EncodeToString(pkBytes),
			"privateKey": base64.StdEncoding.EncodeToString(skBytes),
		}
	})

	// Expose the `export` function to JavaScript
	js.Global().Set("exportKeypair", export)

	// Keep the Wasm module alive
	select {}
}
