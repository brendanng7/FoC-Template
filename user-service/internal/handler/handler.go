// Package handler adapts HTTP requests to user-service operations.
package handler

import "net/http"

// Handler owns the service's HTTP routes.
type Handler struct {
	Router *http.ServeMux
}

// New creates the public service endpoints. Account routes will be registered
// here as their application workflows are implemented.
func New() *Handler {
	router := http.NewServeMux()
	router.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	router.HandleFunc("GET /public", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{\"service\":\"user-service\"}\n"))
	})
	return &Handler{Router: router}
}
