package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthWithoutTokenAllowsFlags(t *testing.T) {
	t.Setenv("API_TOKEN", "")
	api := NewAPI()

	req := httptest.NewRequest(http.MethodGet, "/flags", nil)
	rr := httptest.NewRecorder()
	api.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 without token, got %d", rr.Code)
	}
}

func TestAuthMissingTokenReturns401(t *testing.T) {
	t.Setenv("API_TOKEN", "secret")
	api := NewAPI()

	for _, method := range []string{http.MethodPost, http.MethodGet} {
		req := httptest.NewRequest(method, "/flags", nil)
		rr := httptest.NewRecorder()
		api.routes().ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("%s /flags: expected 401, got %d", method, rr.Code)
		}
		var body map[string]string
		if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
			t.Fatalf("invalid json: %v", err)
		}
		if body["error"] != "unauthorized" {
			t.Fatalf("expected 'unauthorized', got %q", body["error"])
		}
	}
}

func TestAuthWrongTokenReturns401(t *testing.T) {
	t.Setenv("API_TOKEN", "secret")
	api := NewAPI()

	req := httptest.NewRequest(http.MethodGet, "/flags", nil)
	req.Header.Set("Authorization", "Bearer wrong")
	rr := httptest.NewRecorder()
	api.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestAuthCorrectTokenReturns2xx(t *testing.T) {
	t.Setenv("API_TOKEN", "secret")
	api := NewAPI()

	req := httptest.NewRequest(http.MethodGet, "/flags", nil)
	req.Header.Set("Authorization", "Bearer secret")
	rr := httptest.NewRecorder()
	api.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestHealthzReachableWithoutToken(t *testing.T) {
	t.Setenv("API_TOKEN", "secret")
	api := NewAPI()

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	api.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}
