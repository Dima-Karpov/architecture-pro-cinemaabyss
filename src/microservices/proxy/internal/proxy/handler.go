package proxy

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/cinemaabyss/microservices/proxy/internal/api"
	"github.com/cinemaabyss/microservices/proxy/internal/api/responses"
	"github.com/cinemaabyss/microservices/proxy/internal/config"
	"github.com/cinemaabyss/microservices/proxy/internal/strangler"
)

type Handler struct {
	router   *strangler.Router
	monolith *httputil.ReverseProxy
	events   *httputil.ReverseProxy
	movies   *httputil.ReverseProxy
}

func NewHandler(cfg config.Config) *Handler {
	h := &Handler{
		router: strangler.New(cfg),
	}

	h.monolith = newReverseProxy(cfg.MonolithURL, h.handleProxyError)
	h.events = newReverseProxy(cfg.EventsServiceURL, h.handleProxyError)
	h.movies = &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			backend := h.router.SelectMoviesBackend()
			setProxyTarget(req, backend.URL)
		},
		ErrorHandler: h.handleProxyError,
	}

	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if path == "/health" {
		api.HealthHandler(w, r)

		return
	}

	switch {
	case isMoviesPath(path):
		h.movies.ServeHTTP(w, r)
	case isEventsPath(path):
		h.events.ServeHTTP(w, r)
	case isMonolithPath(path):
		h.monolith.ServeHTTP(w, r)
	default:
		responses.WriteError(w, http.StatusNotFound, "NotFoundError", "route not found")
	}
}

func newReverseProxy(target *url.URL, onError func(http.ResponseWriter, *http.Request, error)) *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.ErrorHandler = onError

	return proxy
}

func setProxyTarget(req *http.Request, target *url.URL) {
	req.URL.Scheme = target.Scheme
	req.URL.Host = target.Host
	req.Host = target.Host
}

func (h *Handler) handleProxyError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("proxy error: %s %s: %v", r.Method, r.URL.Path, err)
	responses.WriteError(w, http.StatusBadGateway, "BadGatewayError", "upstream unavailable")
}

func isMoviesPath(path string) bool {
	return path == "/api/movies" || strings.HasPrefix(path, "/api/movies/")
}

func isEventsPath(path string) bool {
	return strings.HasPrefix(path, "/api/events/")
}

func isMonolithPath(path string) bool {
	for _, prefix := range []string{"/api/users", "/api/payments", "/api/subscriptions"} {
		if path == prefix || strings.HasPrefix(path, prefix+"/") {
			return true
		}
	}

	return false
}
