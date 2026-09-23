package config

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	t.Setenv("DATABASE_URL", " ")
	t.Setenv("AUTH0_DOMAIN", "dev-example.us.auth0.com")
	t.Setenv("AUTH0_AUDIENCE", "https://user-service.example.com")
	if _, err := Load(); err == nil {
		t.Fatal("missing database configuration accepted")
	}
	t.Setenv("DATABASE_URL", "postgres://student:secret@localhost/users")
	t.Setenv("HTTP_ADDRESS", "")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.HTTPAddress != ":8080" || cfg.DatabaseURL == "" {
		t.Fatal("configuration not loaded")
	}
	encoded, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "secret") {
		t.Fatal("database credentials serialized")
	}
	t.Setenv("HTTP_ADDRESS", "127.0.0.1:9090")
	cfg, err = Load()
	if err != nil || cfg.HTTPAddress != "127.0.0.1:9090" {
		t.Fatal("HTTP address override not loaded")
	}
}

func TestLoadRequiresAuth0Configuration(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://student:secret@localhost/users")
	t.Setenv("AUTH0_DOMAIN", "")
	t.Setenv("AUTH0_AUDIENCE", "https://user-service.example.com")
	if _, err := Load(); err == nil {
		t.Fatal("missing Auth0 domain accepted")
	}

	t.Setenv("AUTH0_DOMAIN", "dev-example.us.auth0.com")
	t.Setenv("AUTH0_AUDIENCE", "")
	if _, err := Load(); err == nil {
		t.Fatal("missing Auth0 audience accepted")
	}
}
