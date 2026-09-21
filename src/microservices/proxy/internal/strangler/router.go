package strangler

import (
	"log"
	"math/rand"
	"net/url"

	"github.com/cinemaabyss/microservices/proxy/internal/config"
)

type Router struct {
	cfg config.Config
}

type Backend struct {
	URL  *url.URL
	Name string
}

func New(cfg config.Config) *Router {
	return &Router{cfg: cfg}
}

const migrationPercentBase = 100

func (r *Router) SelectMoviesBackend() Backend {
	if !r.cfg.GradualMigration {
		log.Printf("movies route: backend=movies-service (gradual migration disabled)")

		return Backend{Name: "movies-service", URL: r.cfg.MoviesServiceURL}
	}

	if rand.Intn(migrationPercentBase) < r.cfg.MigrationPercent {
		log.Printf("movies route: backend=movies-service")

		return Backend{Name: "movies-service", URL: r.cfg.MoviesServiceURL}
	}

	log.Printf("movies route: backend=monolith")

	return Backend{Name: "monolith", URL: r.cfg.MonolithURL}
}
