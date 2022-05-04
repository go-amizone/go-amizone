package handlers

import (
	"encoding/json"
	"net/http"
)

// WriteJsonResponse encodes the date passed to JSON and writes it to the response along
// with an `application/json` Content-Type header.
func WriteJsonResponse(data interface{}, w http.ResponseWriter) error {
	w.Header().Add("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}
	return nil
}
