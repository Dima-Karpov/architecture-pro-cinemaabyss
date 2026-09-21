package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/cinemaabyss/microservices/events/internal/api/models"
	"github.com/cinemaabyss/microservices/events/internal/api/responses"
	"github.com/cinemaabyss/microservices/events/internal/kafka"
)

// CreateMovieEvent registers a movie-related event and publishes it to Kafka.
//
//	@Summary		Создание события фильма
//	@Description	Регистрирует новое событие, связанное с фильмом
//	@Tags			events
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.MovieEventPayload	true	"Данные события фильма"
//	@Success		201		{object}	models.MovieEventResponse
//	@Failure		400		{object}	responses.ErrorResponse
//	@Failure		500		{object}	responses.ErrorResponse
//	@Router			/api/events/movie [post]
func (h *Handler) CreateMovieEvent(w http.ResponseWriter, r *http.Request) {
	if !requirePost(w, r) {
		return
	}

	var payload models.MovieEventPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		responses.WriteError(w, http.StatusBadRequest, "ValidationError", "invalid request body")

		return
	}

	if payload.MovieID == 0 || !hasText(payload.Title) || !hasText(payload.Action) {
		responses.WriteError(w, http.StatusBadRequest, "ValidationError", "missing required fields")

		return
	}

	event := models.MovieEvent{
		ID:        fmt.Sprintf("movie-%d-%s", payload.MovieID, payload.Action),
		Type:      "movie",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Payload:   payload,
	}

	publishEvent(r.Context(), h, w, kafka.TopicMovieEvents, event,
		func(event models.MovieEvent, result kafka.PublishResult) models.MovieEventResponse {
			return models.MovieEventResponse{
				Status:    eventStatusSuccess,
				Partition: result.Partition,
				Offset:    result.Offset,
				Event:     event,
			}
		},
	)
}
