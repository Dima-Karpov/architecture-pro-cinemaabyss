package responses

import (
	"encoding/json"
	"log"
	"net/http"
)

type Error struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Errors []Error `json:"errors"`
}

func WriteError(w http.ResponseWriter, status int, errType, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(ErrorResponse{
		Errors: []Error{{Type: errType, Message: message}},
	}); err != nil {
		log.Printf("write error response: %v", err)
	}
}
