package handlers

import (
	"encoding/json"
	"net/http"

	"gorm.io/gorm"
)

var DB *gorm.DB

func JSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func JSONResponse(w http.ResponseWriter, status int, value interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(value)
}

func DecodeJSON(w http.ResponseWriter, r *http.Request, value interface{}) bool {
	if err := json.NewDecoder(r.Body).Decode(value); err != nil {
		JSONError(w, http.StatusBadRequest, "Invalid JSON body")
		return false
	}
	return true
}
