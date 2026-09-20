package kafka

const (
	TopicMovieEvents   = "movie-events"
	TopicUserEvents    = "user-events"
	TopicPaymentEvents = "payment-events"
)

func Topics() []string {
	return []string{
		TopicMovieEvents,
		TopicUserEvents,
		TopicPaymentEvents,
	}
}
