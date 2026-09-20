// @title           CinemaAbyss Events Service API
// @version         1.0
// @description     POST endpoints for movie, user, and payment events (Kafka ingestion).
// @host            localhost:8082
// @BasePath        /
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpSwagger "github.com/swaggo/http-swagger"

	_ "github.com/cinemaabyss/microservices/events/docs"
	"github.com/cinemaabyss/microservices/events/internal/api"
	"github.com/cinemaabyss/microservices/events/internal/api/responses"
	"github.com/cinemaabyss/microservices/events/internal/config"
	"github.com/cinemaabyss/microservices/events/internal/kafka"
)

const (
	readHeaderTimeout = 5 * time.Second
	shutdownTimeout   = 10 * time.Second
	consumerGroupID   = "events-service"
)

func main() {
	cfg := config.Load()

	if cfg.KafkaBrokers == "" {
		log.Fatal("KAFKA_BROKERS is required")
	}

	if err := run(cfg); err != nil {
		log.Fatal(err)
	}
}

func run(cfg config.Config) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	consumer := kafka.NewConsumer(cfg.KafkaBrokers, consumerGroupID)
	consumer.Start(ctx)

	producer := kafka.NewProducer(cfg.KafkaBrokers)

	defer closeKafka(stop, consumer, producer)

	srv := newServer(cfg, api.NewHandler(producer))
	errCh := make(chan error, 1)

	go func() {
		log.Printf("Starting events microservice on port %s", cfg.Port)

		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("http server: %w", err)
		}
	case <-ctx.Done():
		log.Printf("shutting down events microservice")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("http shutdown: %w", err)
	}

	return nil
}

func newServer(cfg config.Config, eventHandler *api.Handler) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/swagger/", httpSwagger.WrapHandler)
	mux.HandleFunc("/api/events/health", eventHandler.Health)
	mux.HandleFunc("/api/events/movie", eventHandler.CreateMovieEvent)
	mux.HandleFunc("/api/events/user", eventHandler.CreateUserEvent)
	mux.HandleFunc("/api/events/payment", eventHandler.CreatePaymentEvent)
	mux.HandleFunc("/", notFoundHandler)

	return &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           mux,
		ReadHeaderTimeout: readHeaderTimeout,
	}
}

func closeKafka(stop context.CancelFunc, consumer *kafka.Consumer, producer *kafka.Producer) {
	stop()

	if err := consumer.Close(); err != nil {
		log.Printf("close kafka consumer: %v", err)
	}

	if err := producer.Close(); err != nil {
		log.Printf("close kafka producer: %v", err)
	}
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	responses.WriteError(w, http.StatusNotFound, "NotFoundError", "route not found")
}
