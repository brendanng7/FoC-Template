package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPublicRoutes(t *testing.T) {
	handler := New()

	tests := []struct {
		path       string
		wantStatus int
	}{
		{path: "/health", wantStatus: http.StatusNoContent},
		{path: "/public", wantStatus: http.StatusOK},
	}

	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			response := httptest.NewRecorder()
			handler.Router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, test.path, nil))
			if response.Code != test.wantStatus {
				t.Fatalf("got status %d, want %d", response.Code, test.wantStatus)
			}
		})
	}
}
