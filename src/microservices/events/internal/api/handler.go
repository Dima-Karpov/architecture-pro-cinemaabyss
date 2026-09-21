package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/cinemaabyss/microservices/events/internal/api/responses"
	"github.com/cinemaabyss/microservices/events/internal/kafka"
)

const eventStatusSuccess = "success"

type Handler struct {
	producer *kafka.Producer
}

func NewHandler(producer *kafka.Producer) *Handler {
	return &Handler{producer: producer}
}

func publishEvent[T any, R any](
	ctx context.Context,
	h *Handler,
	w http.ResponseWriter,
	topic string,
	event T,
	buildResponse func(T, kafka.PublishResult) R,
) {
	payload, err := json.Marshal(event)
	if err != nil {
		responses.WriteError(w, http.StatusInternalServerError, "InternalError", "failed to serialize event")

		return
	}

	result, err := h.producer.Publish(ctx, topic, payload)
	if err != nil {
		log.Printf("kafka publish error: topic=%s err=%v", topic, err)
		responses.WriteError(w, http.StatusInternalServerError, "InternalError", "failed to publish event")

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err = json.NewEncoder(w).Encode(buildResponse(event, result)); err != nil {
		log.Printf("write event response: topic=%s err=%v", topic, err)
	}
}

func requirePost(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)

		return false
	}

	return true
}

func hasText(value string) bool {
	return value != ""
}
