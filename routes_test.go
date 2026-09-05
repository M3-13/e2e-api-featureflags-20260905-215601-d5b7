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

	if err := api.store.Put(Flag{Key: "foo", Enabled: true}); err != nil {
		t.Fatalf("failed to seed flag: %v", err)
	}

	assertStatus := func(t *testing.T, rr *httptest.ResponseRecorder, status int) {
		t.Helper()
		if rr.Code != status {
			t.Fatalf("expected %d, got %d: %s", status, rr.Code, rr.Body.String())
		}
	}

	assertError := func(t *testing.T, rr *httptest.ResponseRecorder, status int, wantMsg string) {
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
		assertStatus(t, rr, http.StatusOK)
		var body Flag
		if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
			t.Fatalf("invalid json: %v", err)
		}
		if body.Key != "foo" {
			t.Fatalf("expected key %q, got %q", "foo", body.Key)
		}
	})

	t.Run("flag_put", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/flags/foo", strings.NewReader(`{"enabled":false,"rollout_percent":0}`))
		rr := httptest.NewRecorder()
		api.routes().ServeHTTP(rr, req)
		assertStatus(t, rr, http.StatusOK)
		var body Flag
		if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
			t.Fatalf("invalid json: %v", err)
		}
		if body.Enabled {
			t.Fatalf("expected enabled=false, got true")
		}
	})

	t.Run("flag_evaluate", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/flags/foo/evaluate?user=42", nil)
		rr := httptest.NewRecorder()
		api.routes().ServeHTTP(rr, req)
		assertStatus(t, rr, http.StatusOK)
		var body map[string]bool
		if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
			t.Fatalf("invalid json: %v", err)
		}
		if _, ok := body["enabled"]; !ok {
			t.Fatalf("expected enabled field in response, got %s", rr.Body.String())
		}
	})

	t.Run("flag_delete", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/flags/foo", nil)
		rr := httptest.NewRecorder()
		api.routes().ServeHTTP(rr, req)
		assertStatus(t, rr, http.StatusNoContent)

		req = httptest.NewRequest(http.MethodGet, "/flags/foo", nil)
		rr = httptest.NewRecorder()
		api.routes().ServeHTTP(rr, req)
		assertError(t, rr, http.StatusNotFound, "flag not found")
	})

	t.Run("catchall_not_found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/does-not-exist", nil)
		rr := httptest.NewRecorder()
		api.routes().ServeHTTP(rr, req)
		assertError(t, rr, http.StatusNotFound, "not found")
	})
}
