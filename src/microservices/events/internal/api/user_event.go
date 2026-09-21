package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cinemaabyss/microservices/events/internal/api/models"
	"github.com/cinemaabyss/microservices/events/internal/api/responses"
	"github.com/cinemaabyss/microservices/events/internal/kafka"
)

// CreateUserEvent registers a user-related event and publishes it to Kafka.
//
//	@Summary		Создание события пользователя
//	@Description	Регистрирует новое событие, связанное с пользователем
//	@Tags			events
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.UserEventPayload	true	"Данные события пользователя"
//	@Success		201		{object}	models.UserEventResponse
//	@Failure		400		{object}	responses.ErrorResponse
//	@Failure		500		{object}	responses.ErrorResponse
//	@Router			/api/events/user [post]
func (h *Handler) CreateUserEvent(w http.ResponseWriter, r *http.Request) {
	if !requirePost(w, r) {
		return
	}

	var payload models.UserEventPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		responses.WriteError(w, http.StatusBadRequest, "ValidationError", "invalid request body")

		return
	}

	if payload.UserID == 0 || !hasText(payload.Action) || !hasText(payload.Timestamp) {
		responses.WriteError(w, http.StatusBadRequest, "ValidationError", "missing required fields")

		return
	}

	event := models.UserEvent{
		ID:        fmt.Sprintf("user-%d-%s", payload.UserID, payload.Action),
		Type:      "user",
		Timestamp: payload.Timestamp,
		Payload:   payload,
	}

	publishEvent(r.Context(), h, w, kafka.TopicUserEvents, event,
		func(event models.UserEvent, result kafka.PublishResult) models.UserEventResponse {
			return models.UserEventResponse{
				Status:    eventStatusSuccess,
				Partition: result.Partition,
				Offset:    result.Offset,
				Event:     event,
			}
		},
	)
}
