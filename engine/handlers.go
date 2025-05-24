package main

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
)

var def_parameter = `{"LogN":12,"LogQ":[58],"PlaintextModulus": 65537}`

func (calc *calculator) home(w http.ResponseWriter, r *http.Request) {
	// Use the new render helper.
	w.Write([]byte("heyho"))
}

type CreateStreamRequest struct {
	PK string `json:"pk"`
}

type CreateStreamResponse struct {
	Message string `json:"message"`
	ID      string `json:"id"`
}

// @Summary		Create a new stream
// @Description	Create a new stream with the provided public key
// @Tags			BasicAPI
// @Accept			json
// @Produce		json
// @Param			pk	body		string	true	"BASE64 encoded Public Key"
// @Success		200	{object}	CreateStreamResponse
// @Failure		500
// @Router			/create-stream [post]
// @Security		APIKeyAuth
func (calc *calculator) createStream(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		calc.serverError(w, r, err)
		return
	}
	// Parse the JSON to extract the publicKey field
	var payload map[string]string
	if err := json.Unmarshal(body, &payload); err != nil {
		calc.serverError(w, r, err)
		return
	}
	publicKeyBase64 := payload["pk"]
	// Decode the base64-encoded public key
	publicKeyBytes, err := base64.StdEncoding.DecodeString(publicKeyBase64)
	if err != nil {
		calc.serverError(w, r, err)
		return
	}
	id := generateFreshID()
	err = calc.calc_model.InsertAggregationParams(id, publicKeyBytes, def_parameter)
	if err != nil {
		calc.serverError(w, r, err)
		return
	}
	agg, err := calc.initAggregator(publicKeyBytes, def_parameter)
	if err != nil {
		calc.serverError(w, r, err)
		return
	}
	calc.agg_map.Store(id, agg)

	response := map[string]string{"message": "Token Valid", "id": id}
	writeJSON(w, http.StatusOK, response)
}

type ContributeAggregateRequest struct {
	ID string `json:"id"`
	CT string `json:"ct"`
}

type ConttributeAggregateResponse struct {
	Message string `json:"message"`
	ID      string `json:"id"`
}

// @Summary		Contribute to an existing aggregate
// @Description	Contribute data to an existing aggregate by ID
// @Tags			BasicAPI
// @Accept			json
// @Produce		json
// @Param			id	body		string	true	"Stream ID"
// @Param			ct	body		string	true	"BASE64 encoded Ciphertext"
// @Success		200	{object}	ConttributeAggregateResponse
// @Failure		500
// @Router			/contribute/aggregate/{id} [post]
func (calc *calculator) contributeAggregate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	flag, err := calc.calc_model.IDexists(id)
	if err != nil {
		calc.serverError(w, r, err)
		return
	}
	if !flag {
		w.Write([]byte("Your ID is invalid."))
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		calc.serverError(w, r, err)
		return
	}
	var payload map[string]string
	if err := json.Unmarshal(body, &payload); err != nil {
		calc.serverError(w, r, err)
		return
	}
	ct_base64 := payload["ct"]
	ct, err := base64.StdEncoding.DecodeString(ct_base64)
	if err != nil {
		calc.serverError(w, r, err)
		return
	}
	var agg *aggregator
	if value, ok := calc.agg_map.Load(id); ok {
		agg = value.(*aggregator)
	} else {
		// Initialize aggregator if not already present.
		_, pkBytes, params, err := calc.calc_model.GetAggregationParamsByID(id)
		if err != nil {
			calc.serverError(w, r, err)
			return
		}
		agg, err = calc.initAggregator(pkBytes, params)
		if err != nil {
			calc.serverError(w, r, err)
			return
		}
		calc.agg_map.Store(id, agg)
	}
	err = agg.aggregate(ct)
	if err != nil {
		calc.serverError(w, r, err)
		return
	}
	response := map[string]string{"message": "Input Contributed", "id": id}
	writeJSON(w, http.StatusOK, response)
}

type ReturnAggregateRequest struct {
	ID string `json:"id"`
}

type ReturnAggregateNoneAvailableResponse struct {
	Message string `json:"message"`
	ID      string `json:"id"`
}

type ReturnAggregateResponse struct {
	ID                  string `json:"id"`
	Ct_aggr_byte_base64 string `json:"ct_aggr_byte_base64"`
	SampleSize          int    `json:"sample_size"`
}

// @Summary		Make a snapshot of an existing aggregate
// @Description	Make a snapshot of an existing aggregate by ID
// @Tags			BasicAPI
// @Accept			json
// @Produce		json
// @Param			id	body		string	true	"Stream ID"
// @Success		200	{object}	ReturnAggregateResponse
// @Success		222	{object}	ReturnAggregateNoneAvailableResponse "Request successful, but no aggregate available. Try again later."
// @Failure		500
// @Router			/snapshot/aggregate/{id} [get]
func (calc *calculator) returnAggregate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var agg *aggregator
	if value, ok := calc.agg_map.Load(id); ok {
		agg = value.(*aggregator)
	} else {
		response := &ReturnAggregateNoneAvailableResponse{Message: "Request successful, but no aggregate available. Try again later.", ID: id}
		writeJSON(w, 222, response)
		return
	}
	if agg.ct_aggr == nil {
		response := &ReturnAggregateNoneAvailableResponse{Message: "Request successful, but no aggregate available. Try again later.", ID: id}
		writeJSON(w, 222, response)
		return
	}
	ct_aggr_byte := agg.snapshotAggregate()
	err := calc.calc_model.InsertAggregation(id, ct_aggr_byte, agg.ctr)
	if err != nil {
		calc.serverError(w, r, err)
		return
	}
	base64_ct := base64.StdEncoding.EncodeToString(ct_aggr_byte)
	response := map[string]interface{}{
		"id":                  id,
		"ct_aggr_byte_base64": base64_ct,
		"sample_size":         agg.ctr,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		calc.serverError(w, r, err)
	}
}

type GetPublicKeyResponse struct {
	ID        string `json:"id"`
	PublicKey string `json:"publicKey"`
}

// @Summary		Retrieve the public key associated with a given ID
// @Description	Retrieve the public key (base64-encoded) associated with the given ID.
// @Tags			BasicAPI
// @Accept			json
// @Produce		json
// @Param			id	body		string	true	"Stream ID"
// @Success		200	{object}	GetPublicKeyResponse
// @Failure		500
// @Router			/public-key/{id} [get]
func (calc *calculator) getPublicKey(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	retrievedID, pkBytes, _, err := calc.calc_model.GetAggregationParamsByID(id)
	if err != nil {
		calc.serverError(w, r, err)
		return
	}
	if retrievedID == "" {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "ID not found"})
		calc.serverError(w, r, err)
		return
	}
	pkBase64 := base64.StdEncoding.EncodeToString(pkBytes)
	response := map[string]string{"id": retrievedID, "publicKey": pkBase64}
	writeJSON(w, http.StatusOK, response)
}
