package utils

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"study-planner/dto"
)

const maxJSONBodySize = 1 << 20

func JSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, dto.ErrorResponse{Error: message})
}

func JSONResponse(w http.ResponseWriter, status int, value interface{}) {
	writeJSON(w, status, value)
}

func writeJSON(w http.ResponseWriter, status int, value interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("Could not encode JSON response: %v", err)
	}
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
