package validator

import (
	"fmt"
	"unicode/utf8"
)

// AuxSchema lists the *only* top-level keys we allow in aux.
// Extend as needed.
var AuxSchema = map[string]struct{}{
	"qst":     {},
	"answrs":  {},
	"options": {},
}

// ValidateAux takes raw JSON (after you've decoded req.Aux) and
// populates v.FieldErrors if anything is off.
//
// Typical call in your handler:
//
//	var req CreateStreamRequest
//	err := json.NewDecoder(r.Body).Decode(&req)
//	...
//	var aux map[string]any
//	if len(req.Aux) > 0 {
//	    if err := json.Unmarshal(req.Aux, &aux); err != nil {
//	         v.AddFieldError("aux", "must be valid JSON")
//	    } else {
//	         ValidateAux(v, aux)
//	    }
//	}
func ValidateAux(v *Validator, aux interface{}) {
	switch val := aux.(type) {
	case map[string]any:
		// Existing object validation logic
		// 1) Unknown keys?
		for k := range val {
			if _, ok := AuxSchema[k]; !ok {
				v.AddFieldError("aux."+k, "unexpected field")
			}
		}

		// 2) qst must be []string, each 1-200 runes
		if q, ok := val["qst"]; ok {
			arr, ok := q.([]any)
			if !ok {
				v.AddFieldError("aux.qst", "must be an array")
			} else {
				for i, raw := range arr {
					s, ok := raw.(string)
					if !ok {
						v.AddFieldError(fmt.Sprintf("aux.qst[%d]", i), "must be string")
						continue
					}
					if utf8.RuneCountInString(s) == 0 {
						v.AddFieldError(fmt.Sprintf("aux.qst[%d]", i), "must not be blank")
					}
					if utf8.RuneCountInString(s) > 200 {
						v.AddFieldError(fmt.Sprintf("aux.qst[%d]", i), "must be ≤ 200 chars")
					}
				}
			}
		}

		// 3) answrs must be []string or [][]string depending on your design.
		// Example: optional array of allowed answers (flat).
		if a, ok := val["answrs"]; ok {
			arr, ok := a.([]any)
			if !ok {
				v.AddFieldError("aux.answrs", "must be an array")
			} else {
				for i, raw := range arr {
					s, ok := raw.(string)
					if !ok {
						v.AddFieldError(fmt.Sprintf("aux.answrs[%d]", i), "must be string")
						continue
					}
					if utf8.RuneCountInString(s) > 100 {
						v.AddFieldError(fmt.Sprintf("aux.answrs[%d]", i), "must be ≤ 100 chars")
					}
				}
			}
		}

		// 4) You can add further checks, e.g. options must be an object with bool/int values
	case []any:
		// New: Validate each object in the array
		for i, item := range val {
			obj, ok := item.(map[string]any)
			if !ok {
				v.AddFieldError(fmt.Sprintf("aux[%d]", i), "must be an object")
				continue
			}
			// Recursively validate each object
			ValidateAux(v, obj)
		}
	default:
		v.AddFieldError("aux", "must be a JSON object or array of objects")
	}
}
