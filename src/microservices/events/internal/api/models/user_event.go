package models

type UserEventPayload struct {
	Username  string `json:"username,omitempty"`
	Email     string `json:"email,omitempty"`
	Action    string `json:"action"`
	Timestamp string `json:"timestamp"`
	UserID    int    `json:"user_id"`
}

type UserEvent struct {
	ID        string           `json:"id"`
	Type      string           `json:"type"`
	Timestamp string           `json:"timestamp"`
	Payload   UserEventPayload `json:"payload"`
}

type UserEventResponse struct {
	Status    string    `json:"status"`
	Event     UserEvent `json:"event"`
	Offset    int64     `json:"offset"`
	Partition int       `json:"partition"`
}
