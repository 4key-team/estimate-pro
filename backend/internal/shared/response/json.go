package response

import (
	"encoding/json"
	"net/http"
)

// WriteJSON encodes v as JSON and writes it to the response writer with the given status code.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
