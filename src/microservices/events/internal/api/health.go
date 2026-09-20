package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

const healthCheckTimeout = 2 * time.Second

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)

		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), healthCheckTimeout)
	defer cancel()

	if err := h.producer.Ping(ctx); err != nil {
		log.Printf("kafka health check failed: %v", err)
		writeHealth(w, http.StatusServiceUnavailable, false)

		return
	}

	writeHealth(w, http.StatusOK, true)
}

func writeHealth(w http.ResponseWriter, statusCode int, ok bool) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(map[string]bool{"status": ok}); err != nil {
		log.Printf("write health response: %v", err)
	}
}
