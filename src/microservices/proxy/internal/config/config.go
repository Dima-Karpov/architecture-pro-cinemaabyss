package config

import (
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	MonolithURL      *url.URL
	MoviesServiceURL *url.URL
	EventsServiceURL *url.URL
	Port             string
	MigrationPercent int
	GradualMigration bool
}

const maxMigrationPercent = 100

func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	gradualMigration := strings.ToLower(os.Getenv("GRADUAL_MIGRATION")) == "true"

	migrationPercent := 50

	if pct := os.Getenv("MOVIES_MIGRATION_PERCENT"); pct != "" {
		if n, err := strconv.Atoi(pct); err != nil {
			log.Printf("invalid MOVIES_MIGRATION_PERCENT %q, using default 50", pct)
		} else {
			migrationPercent = clampPercent(n)
		}
	}

	return Config{
		Port:             port,
		MonolithURL:      mustParseURL(envOrDefault("MONOLITH_URL", "http://localhost:8080")),
		MoviesServiceURL: mustParseURL(envOrDefault("MOVIES_SERVICE_URL", "http://localhost:8081")),
		EventsServiceURL: mustParseURL(envOrDefault("EVENTS_SERVICE_URL", "http://localhost:8082")),
		GradualMigration: gradualMigration,
		MigrationPercent: migrationPercent,
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}

func mustParseURL(raw string) *url.URL {
	parsed, err := url.Parse(raw)
	if err != nil {
		log.Fatalf("invalid URL %q: %v", raw, err)
	}

	return parsed
}

func clampPercent(n int) int {
	if n < 0 {
		return 0
	}

	if n > maxMigrationPercent {
		return maxMigrationPercent
	}

	return n
}
