package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"

	"github.com/collapsinghierarchy/encproc/validator"
)

var def_parameter = `{"LogN":12,"LogQ":[58],"PlaintextModulus": 65537}`
var thumbsUpCount int64 // global variable, or use sync/atomic for concurrency

type CreateStreamRequest struct {
	PK  string          `json:"pk"`
	Aux json.RawMessage `json:"aux,omitempty"`
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
	var req CreateStreamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		calc.serverError(w, r, err) // bad JSON → 400/500
		return
	}
	v := &validator.Validator{}

	// ── validate pk ────────────────────────────────────────────────
	if !validator.NotBlank(req.PK) {
		v.AddFieldError("pk", "must be provided")
	}

	// ── validate aux (if supplied) ────────────────────────────────
	if len(req.Aux) > 0 {
		fmt.Print("createStream: aux=", string(req.Aux), "\n")
		// TODO: No validation, just store as-is. Potentially dangerous. (XSS vulnerability)
	}

	if !v.Valid() {
		response := map[string]interface{}{
			"message": "Invalid request",
			"errors":  v.FieldErrors,
		}
		writeJSON(w, http.StatusBadRequest, response)
		return
	}
	id := generateFreshID()
	calc.aux_map.Store(id, req.Aux)
	publicKeyBase64 := req.PK
	if publicKeyBase64 == "" {
		response := map[string]string{"error": "Public key must be provided"}
		writeJSON(w, http.StatusBadRequest, response)
		return
	}
	// Decode the base64-encoded public key
	publicKeyBytes, err := base64.StdEncoding.DecodeString(publicKeyBase64)
	if err != nil {
		calc.serverError(w, r, err)
		return
	}
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
// @Router			/contribute/aggregate [post]
func (calc *calculator) contributeAggregate(w http.ResponseWriter, r *http.Request) {
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
	id := payload["id"]
	fmt.Print("contributeAggregate: id=", id, "\n")
	flag, err := calc.calc_model.IDexists(id)
	if err != nil {
		calc.serverError(w, r, err)
		return
	}
	if !flag {
		response := &ReturnAggregateNoneAvailableResponse{Message: "Your ID is invalid!", ID: id}
		writeJSON(w, 222, response)
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
		fmt.Printf("agg_map.Load: id=%s, found=%v\n", id, ok)
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
		writeJSON(w, 221, response)
		return
	}
	if agg.ct_aggr == nil {
		response := &ReturnAggregateNoneAvailableResponse{Message: "Request successful, but no aggregate available. Try again later.", ID: id}
		writeJSON(w, 222, response)
		return
	}
	ct_aggr_byte, err := agg.snapshotAggregate()
	if err != nil {
		response := &ReturnAggregateNoneAvailableResponse{Message: "Request successful, but no aggregate available. Try again later.", ID: id}
		writeJSON(w, 223, response)
		return
	}
	err = calc.calc_model.InsertAggregation(id, ct_aggr_byte, agg.ctr)
	if err != nil {
		calc.serverError(w, r, err)
		return
	}
	auxVal, ok := calc.aux_map.Load(id)
	if !ok {
		//TODO: acutally its a panic situation...
		response := &ReturnAggregateNoneAvailableResponse{Message: "No Aux data found", ID: id}
		writeJSON(w, 221, response)
		return
	}
	base64_ct := base64.StdEncoding.EncodeToString(ct_aggr_byte)
	response := map[string]interface{}{
		"id":                  id,
		"ct_aggr_byte_base64": base64_ct,
		"sample_size":         agg.ctr,
		"aux":                 auxVal, // value is json.RawMessage, so this will be valid JSON
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
		response := &ReturnAggregateNoneAvailableResponse{Message: "ID not found", ID: id}
		writeJSON(w, 221, response)
		return
	}
	//	var value json.RawMessage
	auxVal, ok := calc.aux_map.Load(id)
	if !ok {
		//TODO: acutally its a panic situation...
		response := &ReturnAggregateNoneAvailableResponse{Message: "No Aux data found", ID: id}
		writeJSON(w, 221, response)
		return
	}
	pkBase64 := base64.StdEncoding.EncodeToString(pkBytes)
	response := map[string]interface{}{
		"id":        retrievedID,
		"publicKey": pkBase64,
		"aux":       auxVal, // value is json.RawMessage, so this will be valid JSON
	}
	writeJSON(w, http.StatusOK, response)
}

// streamDetails serves stream info or the participate page depending on {display}
func (calc *calculator) streamDetails(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	display := r.PathValue("display") // will be "" if not present

	if display == "contribute" {
		//check if aux data is available
		_, ok := calc.aux_map.Load(id)
		if !ok {
			response := &ReturnAggregateNoneAvailableResponse{Message: "No Aux data found", ID: id}
			writeJSON(w, 221, response)
			return
		}
		// Serve participate.html
		http.ServeFile(w, r, "./static/participate.html")
		return
	}

	// Otherwise, return stream details as JSON (example)
	// You can customize this to return whatever stream info you want
	_, pkBytes, _, err := calc.calc_model.GetAggregationParamsByID(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Stream not found"})
		return
	}
	pkBase64 := base64.StdEncoding.EncodeToString(pkBytes)
	response := map[string]interface{}{
		"id":         id,
		"publicKey":  pkBase64,
		"parameters": def_parameter,
		// Add more stream details here as needed
	}
	writeJSON(w, http.StatusOK, response)
}

// Route handler to get the current thumbs up count
func (calc *calculator) getThumbsUp(w http.ResponseWriter, r *http.Request) {
	count := atomic.LoadInt64(&calc.thumbsUpCount)
	writeJSON(w, http.StatusOK, map[string]int64{"count": count})
}

// Route handler to increment the thumbs up count
func (calc *calculator) incrementThumbsUp(w http.ResponseWriter, r *http.Request) {
	newCount := atomic.AddInt64(&calc.thumbsUpCount, 1)
	writeJSON(w, http.StatusOK, map[string]int64{"count": newCount})
}
