package models

type MovieEventPayload struct {
	UserID      *int     `json:"user_id,omitempty"`
	Rating      *float64 `json:"rating,omitempty"`
	Title       string   `json:"title"`
	Action      string   `json:"action"`
	Description string   `json:"description,omitempty"`
	Genres      []string `json:"genres,omitempty"`
	MovieID     int      `json:"movie_id"`
}

type MovieEvent struct {
	ID        string            `json:"id"`
	Type      string            `json:"type"`
	Timestamp string            `json:"timestamp"`
	Payload   MovieEventPayload `json:"payload"`
}

type MovieEventResponse struct {
	Status    string     `json:"status"`
	Event     MovieEvent `json:"event"`
	Offset    int64      `json:"offset"`
	Partition int        `json:"partition"`
}
