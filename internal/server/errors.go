package server

import (
	"encoding/json"
	"net/http"
)

func handleBadRequest(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(newErrorResponse(err.Error()))
}
