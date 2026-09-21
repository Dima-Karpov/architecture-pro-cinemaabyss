package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cinemaabyss/microservices/events/internal/api/models"
	"github.com/cinemaabyss/microservices/events/internal/api/responses"
	"github.com/cinemaabyss/microservices/events/internal/kafka"
)

// CreatePaymentEvent registers a payment-related event and publishes it to Kafka.
//
//	@Summary		Создание события платежа
//	@Description	Регистрирует новое событие, связанное с платежом
//	@Tags			events
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.PaymentEventPayload	true	"Данные события платежа"
//	@Success		201		{object}	models.PaymentEventResponse
//	@Failure		400		{object}	responses.ErrorResponse
//	@Failure		500		{object}	responses.ErrorResponse
//	@Router			/api/events/payment [post]
func (h *Handler) CreatePaymentEvent(w http.ResponseWriter, r *http.Request) {
	if !requirePost(w, r) {
		return
	}

	var payload models.PaymentEventPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		responses.WriteError(w, http.StatusBadRequest, "ValidationError", "invalid request body")

		return
	}

	if payload.PaymentID == 0 || payload.UserID == 0 || !hasText(payload.Status) || !hasText(payload.Timestamp) {
		responses.WriteError(w, http.StatusBadRequest, "ValidationError", "missing required fields")

		return
	}

	event := models.PaymentEvent{
		ID:        fmt.Sprintf("payment-%d-%s", payload.PaymentID, payload.Status),
		Type:      "payment",
		Timestamp: payload.Timestamp,
		Payload:   payload,
	}

	publishEvent(r.Context(), h, w, kafka.TopicPaymentEvents, event,
		func(event models.PaymentEvent, result kafka.PublishResult) models.PaymentEventResponse {
			return models.PaymentEventResponse{
				Status:    eventStatusSuccess,
				Partition: result.Partition,
				Offset:    result.Offset,
				Event:     event,
			}
		},
	)
}
