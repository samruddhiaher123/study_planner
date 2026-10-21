package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"gorm.io/gorm"
)

const maxJSONBodySize = 1 << 20

var DB *gorm.DB

type errorResponse struct {
	Error string `json:"error"`
}

func JSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(errorResponse{Error: message})
}

func JSONResponse(w http.ResponseWriter, status int, value interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(value)
}

func DecodeJSON(w http.ResponseWriter, r *http.Request, value interface{}) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxJSONBodySize)
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(value); err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			JSONError(w, http.StatusRequestEntityTooLarge, "Request body too large")
			return false
		}
		JSONError(w, http.StatusBadRequest, "Invalid JSON body")
		return false
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			JSONError(w, http.StatusRequestEntityTooLarge, "Request body too large")
			return false
		}
		JSONError(w, http.StatusBadRequest, "Invalid JSON body")
		return false
	}
	return true
}
