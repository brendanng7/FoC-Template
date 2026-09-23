// Package middleware authenticates and authorizes HTTP requests.
package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v3"
	"github.com/auth0/go-jwt-middleware/v3/jwks"
	"github.com/auth0/go-jwt-middleware/v3/validator"
)

// Authentication will authenticate a request before invoking the next handler.
type Authentication func(next http.Handler) http.Handler

// Authorization restricts a handler to the specified Auth0 permissions.
type Authorization func(permissions ...string) func(next http.Handler) http.Handler

// CustomClaims contains the Auth0 authorization claims used by this service.
type CustomClaims struct {
	Scope       string   `json:"scope"`
	Permissions []string `json:"permissions"`
}

// Validate satisfies validator.CustomClaims. Authorization is enforced by
// RequirePermissions so tokens without permissions can still reach routes that
// only require authentication.
func (*CustomClaims) Validate(context.Context) error { return nil }

// HasPermission reports whether the token contains the requested permission.
func (c *CustomClaims) HasPermission(permission string) bool {
	for _, granted := range c.Permissions {
		if granted == permission {
			return true
		}
	}
	return false
}

// Auth0 validates RS256 access tokens issued by one Auth0 tenant.
type Auth0 struct {
	middleware *jwtmiddleware.JWTMiddleware
}

// NewAuth0 constructs authentication for the configured tenant and audience.
// excludedPaths are public exact paths, such as /health and /public.
func NewAuth0(domain, audience string, excludedPaths []string) (*Auth0, error) {
	issuerURL, err := parseIssuerURL(domain)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(audience) == "" {
		return nil, errors.New("Auth0 audience is required")
	}

	provider, err := jwks.NewCachingProvider(
		jwks.WithIssuerURL(issuerURL),
		jwks.WithStrictJWKSURIOrigin(),
	)
	if err != nil {
		return nil, fmt.Errorf("create Auth0 JWKS provider: %w", err)
	}

	jwtValidator, err := validator.New(
		validator.WithKeyFunc(provider.KeyFunc),
		validator.WithAlgorithm(validator.RS256),
		validator.WithIssuer(issuerURL.String()),
		validator.WithAudience(strings.TrimSpace(audience)),
		validator.WithCustomClaims(func() *CustomClaims { return &CustomClaims{} }),
	)
	if err != nil {
		return nil, fmt.Errorf("create Auth0 JWT validator: %w", err)
	}

	options := []jwtmiddleware.Option{
		jwtmiddleware.WithValidator(jwtValidator),
		jwtmiddleware.WithErrorHandler(authenticationError),
	}
	if len(excludedPaths) > 0 {
		options = append(options, jwtmiddleware.WithExclusionUrls(excludedPaths))
	}
	jwt, err := jwtmiddleware.New(options...)
	if err != nil {
		return nil, fmt.Errorf("create Auth0 JWT middleware: %w", err)
	}
	return &Auth0{middleware: jwt}, nil
}

func parseIssuerURL(domain string) (*url.URL, error) {
	domain = strings.TrimSpace(domain)
	if domain == "" {
		return nil, errors.New("Auth0 domain is required")
	}
	if strings.Contains(domain, "://") || strings.ContainsAny(domain, "/?#") {
		return nil, errors.New("Auth0 domain must be a hostname without scheme or path")
	}
	issuerURL, err := url.Parse("https://" + domain + "/")
	if err != nil || issuerURL.Hostname() == "" || issuerURL.User != nil {
		return nil, errors.New("Auth0 domain is invalid")
	}
	return issuerURL, nil
}

// Authentication validates a bearer token unless the request path is public.
func (a *Auth0) Authentication(next http.Handler) http.Handler {
	return a.middleware.CheckJWT(next)
}

// RequirePermissions requires every listed permission in the token's
// permissions claim. Apply it inside Authentication.
func (a *Auth0) RequirePermissions(permissions ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, err := jwtmiddleware.GetClaims[*validator.ValidatedClaims](r.Context())
			if err != nil {
				writeJSONError(w, http.StatusUnauthorized, "authentication_required", "A valid access token is required.")
				return
			}
			customClaims, ok := claims.CustomClaims.(*CustomClaims)
			if !ok || customClaims == nil {
				writeJSONError(w, http.StatusForbidden, "insufficient_permissions", "The access token does not contain the required permissions.")
				return
			}
			for _, permission := range permissions {
				if !customClaims.HasPermission(permission) {
					writeJSONError(w, http.StatusForbidden, "insufficient_permissions", "The access token does not contain the required permissions.")
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// Subject returns the authenticated Auth0 subject (the JWT sub claim).
func Subject(ctx context.Context) (string, bool) {
	claims, err := jwtmiddleware.GetClaims[*validator.ValidatedClaims](ctx)
	if err != nil || claims.RegisteredClaims.Subject == "" {
		return "", false
	}
	return claims.RegisteredClaims.Subject, true
}

func authenticationError(w http.ResponseWriter, _ *http.Request, err error) {
	if errors.Is(err, jwtmiddleware.ErrJWTMissing) {
		writeJSONError(w, http.StatusUnauthorized, "missing_token", "Authorization header with a bearer token is required.")
		return
	}
	writeJSONError(w, http.StatusUnauthorized, "invalid_token", "The access token is invalid or expired.")
}

func writeJSONError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": code, "message": message})
}
