package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func evaluate(t *testing.T, api *API, path string) (int, map[string]bool) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rr := httptest.NewRecorder()
	api.routes().ServeHTTP(rr, req)

	var body map[string]bool
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		return rr.Code, nil
	}
	return rr.Code, body
}

func TestEvaluateUserTooLong(t *testing.T) {
	api := NewAPI()
	if err := api.store.Put(Flag{Key: "k", Enabled: true, RolloutPercent: 50}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	longUser := strings.Repeat("a", 257)
	req := httptest.NewRequest(http.MethodGet, "/flags/k/evaluate?user="+longUser, nil)
	rr := httptest.NewRecorder()
	api.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected JSON error body, got %q: %v", rr.Body.String(), err)
	}
	if body["error"] != "user must be at most 256 characters" {
		t.Fatalf("unexpected error message: %v", body)
	}
}

func TestEvaluateUserAtLimit(t *testing.T) {
	api := NewAPI()
	if err := api.store.Put(Flag{Key: "k", Enabled: true, RolloutPercent: 50}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	user := strings.Repeat("a", 256)
	status, _ := evaluate(t, api, "/flags/k/evaluate?user="+user)
	if status != http.StatusOK {
		t.Fatalf("expected 200 for 256-char user, got %d", status)
	}
}

func TestEvaluateMissingUser(t *testing.T) {
	api := NewAPI()
	if err := api.store.Put(Flag{Key: "k", Enabled: true, RolloutPercent: 50}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, path := range []string{"/flags/k/evaluate", "/flags/k/evaluate?user="} {
		status, _ := evaluate(t, api, path)
		if status != http.StatusBadRequest {
			t.Fatalf("expected 400 for %q, got %d", path, status)
		}
	}
}

func TestEvaluateUnknownFlag(t *testing.T) {
	api := NewAPI()
	status, _ := evaluate(t, api, "/flags/nope/evaluate?user=42")
	if status != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", status)
	}
}

func TestEvaluateDecision(t *testing.T) {
	api := NewAPI()
	if err := api.store.Put(Flag{Key: "k", Enabled: true, RolloutPercent: 50}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	status, body := evaluate(t, api, "/flags/k/evaluate?user=42")
	if status != http.StatusOK {
		t.Fatalf("expected 200, got %d", status)
	}
	if len(body) != 1 {
		t.Fatalf("expected exactly one field, got %v", body)
	}
	if _, ok := body["enabled"]; !ok {
		t.Fatalf("expected enabled field, got %v", body)
	}
}

func TestEvaluateDeterministic(t *testing.T) {
	api := NewAPI()
	if err := api.store.Put(Flag{Key: "k", Enabled: true, RolloutPercent: 50}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	first := stableHash("k", "42")%100 < 50
	for i := 0; i < 10; i++ {
		_, body := evaluate(t, api, "/flags/k/evaluate?user=42")
		if body["enabled"] != first {
			t.Fatalf("expected deterministic result %v, got %v", first, body["enabled"])
		}
	}
}

func TestEvaluateRolloutFull(t *testing.T) {
	api := NewAPI()
	if err := api.store.Put(Flag{Key: "k", Enabled: true, RolloutPercent: 100}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	status, body := evaluate(t, api, "/flags/k/evaluate?user=42")
	if status != http.StatusOK {
		t.Fatalf("expected 200, got %d", status)
	}
	if !body["enabled"] {
		t.Fatalf("expected enabled=true for rollout 100")
	}
}

func TestEvaluateRolloutZero(t *testing.T) {
	api := NewAPI()
	if err := api.store.Put(Flag{Key: "k", Enabled: true, RolloutPercent: 0}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	status, body := evaluate(t, api, "/flags/k/evaluate?user=42")
	if status != http.StatusOK {
		t.Fatalf("expected 200, got %d", status)
	}
	if body["enabled"] {
		t.Fatalf("expected enabled=false for rollout 0")
	}
}

func TestEvaluateDisabled(t *testing.T) {
	api := NewAPI()
	if err := api.store.Put(Flag{Key: "k", Enabled: false, RolloutPercent: 100}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	status, body := evaluate(t, api, "/flags/k/evaluate?user=42")
	if status != http.StatusOK {
		t.Fatalf("expected 200, got %d", status)
	}
	if body["enabled"] {
		t.Fatalf("expected enabled=false when flag disabled")
	}
}

func TestEvaluateUserNotStored(t *testing.T) {
	api := NewAPI()
	if err := api.store.Put(Flag{Key: "k", Enabled: true, RolloutPercent: 50}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	user := "secret-user-value"
	_, _ = evaluate(t, api, "/flags/k/evaluate?user="+user)

	for _, f := range api.store.List() {
		if f.Key == user || f.Description == user {
			t.Fatalf("user value leaked into store: %+v", f)
		}
	}
	if _, ok := api.store.Get(user); ok {
		t.Fatalf("user value stored as a flag key")
	}
}
