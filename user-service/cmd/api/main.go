package main

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/joho/godotenv"

	"user-service/internal/config"
	"user-service/internal/handler"
	authmiddleware "user-service/internal/middleware"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// A missing .env is expected in production, where the environment is
	// supplied by the deployment platform.
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	authentication, err := authmiddleware.NewAuth0(
		cfg.Auth0Domain,
		cfg.Auth0Audience,
		[]string{"/health", "/public"},
	)
	if err != nil {
		return err
	}

	httpHandler := handler.New()
	server := &http.Server{
		Addr:              cfg.HTTPAddress,
		Handler:           authentication.Authentication(httpHandler.Router),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("user-service listening on %s", cfg.HTTPAddress)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
