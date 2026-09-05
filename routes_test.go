package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRoutesRegistered(t *testing.T) {
	api := NewAPI()

	assertHandlerError := func(t *testing.T, rr *httptest.ResponseRecorder, status int, wantMsg string) {
		t.Helper()
		if rr.Code != status {
			t.Fatalf("expected %d, got %d: %s", status, rr.Code, rr.Body.String())
		}
		var body map[string]string
		if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
			t.Fatalf("invalid json: %v", err)
		}
		if body["error"] != wantMsg {
			t.Fatalf("expected error %q, got %q", wantMsg, body["error"])
		}
	}

	t.Run("flag_get", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/flags/foo", nil)
		rr := httptest.NewRecorder()
		api.routes().ServeHTTP(rr, req)
		assertHandlerError(t, rr, http.StatusNotFound, "flag not found")
	})

	t.Run("flag_put", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/flags/foo", strings.NewReader(`{"enabled":true}`))
		rr := httptest.NewRecorder()
		api.routes().ServeHTTP(rr, req)
		assertHandlerError(t, rr, http.StatusNotFound, "flag not found")
	})

	t.Run("flag_delete", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/flags/foo", nil)
		rr := httptest.NewRecorder()
		api.routes().ServeHTTP(rr, req)
		assertHandlerError(t, rr, http.StatusNotFound, "flag not found")
	})

	t.Run("flag_evaluate", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/flags/foo/evaluate?user=42", nil)
		rr := httptest.NewRecorder()
		api.routes().ServeHTTP(rr, req)
		assertHandlerError(t, rr, http.StatusNotFound, "flag not found")
	})
}
