package main

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
)

var Validate *validator.Validate

func init() {
	Validate = validator.New(validator.WithRequiredStructEnabled())
}

func RespondWithError(statusCode int, w http.ResponseWriter, message string) {
	type envolope struct {
		Error string `json:"string"`
	}

	RespondWithJson(statusCode, w, &envolope{Error: message})
}

func RespondWithJson(statusCode int, w http.ResponseWriter, payload any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	return json.NewEncoder(w).Encode(payload)
}

func ReadJson(w http.ResponseWriter, r *http.Request, data any) error {
	maxBytes := 1_048_578
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	return decoder.Decode(data)
}

func (app *application) jsonResponse(w http.ResponseWriter, status int, data any) error {
	type envelope struct {
		Data any `json:"data"`
	}

	return RespondWithJson(status, w, &envelope{Data: data})
}
