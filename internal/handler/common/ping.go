package common

import (
	"encoding/json"
	"net/http"
)

func Ping(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"message": "pong"}); err != nil {
		http.Error(w, "encode error", http.StatusInternalServerError)
		return
	}
}
