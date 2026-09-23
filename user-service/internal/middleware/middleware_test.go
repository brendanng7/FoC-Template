package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/auth0/go-jwt-middleware/v3/core"
	"github.com/auth0/go-jwt-middleware/v3/validator"
)

func TestAuthenticationExcludesOnlyConfiguredPublicPaths(t *testing.T) {
	authentication, err := NewAuth0(
		"dev-example.us.auth0.com",
		"https://user-service.example.com",
		[]string{"/health", "/public"},
	)
	if err != nil {
		t.Fatal(err)
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	handler := authentication.Authentication(next)

	for _, path := range []string{"/health", "/public"} {
		t.Run(path, func(t *testing.T) {
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
			if response.Code != http.StatusNoContent {
				t.Fatalf("public path returned %d", response.Code)
			}
		})
	}

	for _, path := range []string{"/", "/profile", "/public/nested"} {
		t.Run(path, func(t *testing.T) {
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
			if response.Code != http.StatusUnauthorized {
				t.Fatalf("protected path returned %d", response.Code)
			}
			if !strings.Contains(response.Body.String(), `"error":"missing_token"`) {
				t.Fatalf("unexpected response: %s", response.Body.String())
			}
		})
	}
}

func TestNewAuth0RejectsInvalidConfiguration(t *testing.T) {
	tests := []struct {
		name     string
		domain   string
		audience string
	}{
		{name: "missing domain", audience: "https://user-service.example.com"},
		{name: "domain with scheme", domain: "https://dev-example.us.auth0.com", audience: "https://user-service.example.com"},
		{name: "domain with path", domain: "dev-example.us.auth0.com/path", audience: "https://user-service.example.com"},
		{name: "missing audience", domain: "dev-example.us.auth0.com"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := NewAuth0(test.domain, test.audience, nil); err == nil {
				t.Fatal("invalid configuration accepted")
			}
		})
	}
}

func TestRequirePermissions(t *testing.T) {
	authentication, err := NewAuth0(
		"dev-example.us.auth0.com",
		"https://user-service.example.com",
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	handler := authentication.RequirePermissions("read:profile", "write:profile")(next)

	tests := []struct {
		name        string
		permissions []string
		wantStatus  int
	}{
		{name: "all permissions", permissions: []string{"read:profile", "write:profile"}, wantStatus: http.StatusNoContent},
		{name: "missing permission", permissions: []string{"read:profile"}, wantStatus: http.StatusForbidden},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			claims := &validator.ValidatedClaims{
				CustomClaims: &CustomClaims{Permissions: test.permissions},
			}
			request := httptest.NewRequest(http.MethodGet, "/profile", nil)
			request = request.WithContext(core.SetClaims(request.Context(), claims))
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != test.wantStatus {
				t.Fatalf("got status %d, want %d", response.Code, test.wantStatus)
			}
		})
	}
}

func TestRequirePermissionsNeedsAuthentication(t *testing.T) {
	authentication, err := NewAuth0(
		"dev-example.us.auth0.com",
		"https://user-service.example.com",
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	handler := authentication.RequirePermissions("read:profile")(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("handler called without authentication")
	}))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/profile", nil))
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("got status %d, want %d", response.Code, http.StatusUnauthorized)
	}
}

func TestSubject(t *testing.T) {
	claims := &validator.ValidatedClaims{
		RegisteredClaims: validator.RegisteredClaims{Subject: "auth0|user-123"},
	}
	ctx := core.SetClaims(context.Background(), claims)
	if subject, ok := Subject(ctx); !ok || subject != "auth0|user-123" {
		t.Fatalf("got subject %q, ok %v", subject, ok)
	}
	if _, ok := Subject(context.Background()); ok {
		t.Fatal("subject returned without claims")
	}
}
