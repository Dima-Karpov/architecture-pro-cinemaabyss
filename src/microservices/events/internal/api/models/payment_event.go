package models

type PaymentEventPayload struct {
	Status     string  `json:"status"`
	Timestamp  string  `json:"timestamp"`
	MethodType string  `json:"method_type,omitempty"`
	Amount     float64 `json:"amount"`
	PaymentID  int     `json:"payment_id"`
	UserID     int     `json:"user_id"`
}

type PaymentEvent struct {
	ID        string              `json:"id"`
	Type      string              `json:"type"`
	Timestamp string              `json:"timestamp"`
	Payload   PaymentEventPayload `json:"payload"`
}

type PaymentEventResponse struct {
	Status    string       `json:"status"`
	Event     PaymentEvent `json:"event"`
	Offset    int64        `json:"offset"`
	Partition int          `json:"partition"`
}
