package api

import (
	"log"
	"net/http"
)

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)

		return
	}

	w.Header().Set("Content-Type", "text/plain")

	if _, err := w.Write([]byte("Strangler Fig Proxy is healthy")); err != nil {
		log.Printf("write health response: %v", err)
	}
}
