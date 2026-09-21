package main

import (
	"log"
	"net/http"
	"time"

	"github.com/cinemaabyss/microservices/proxy/internal/config"
	"github.com/cinemaabyss/microservices/proxy/internal/proxy"
)

const readHeaderTimeout = 5 * time.Second

func main() {
	cfg := config.Load()
	handler := proxy.NewHandler(cfg)

	addr := ":" + cfg.Port
	log.Printf("Starting proxy service on port %s", cfg.Port)
	log.Printf("Gradual migration: %v, movies migration percent: %d", cfg.GradualMigration, cfg.MigrationPercent)

	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: readHeaderTimeout,
	}
	log.Fatal(srv.ListenAndServe())
}
