package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthz(t *testing.T) {
	api := NewAPI()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()

	api.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("expected status ok, got %q", body["status"])
	}
}

func TestHealthzIncludesVersion(t *testing.T) {
	api := NewAPI()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()

	api.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if body["version"] == "" {
		t.Fatalf("expected non-empty version field, got %q", body["version"])
	}
}

func TestNotFound(t *testing.T) {
	api := NewAPI()
	req := httptest.NewRequest(http.MethodGet, "/does-not-exist", nil)
	rr := httptest.NewRecorder()

	api.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
	assertErrorJSON(t, rr)
}

func TestMethodNotAllowed(t *testing.T) {
	api := NewAPI()
	req := httptest.NewRequest(http.MethodDelete, "/healthz", nil)
	rr := httptest.NewRecorder()

	api.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
	assertErrorJSON(t, rr)
}

func assertErrorJSON(t *testing.T, rr *httptest.ResponseRecorder) {
	t.Helper()
	var body map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if msg, ok := body["error"]; !ok || msg == "" {
		t.Fatalf("expected non-empty error message, got %v", body)
	}
	if len(body) != 1 {
		t.Fatalf("error object must contain exactly the error field, got %v", body)
	}
}

func TestWriteError(t *testing.T) {
	rr := httptest.NewRecorder()
	writeError(rr, http.StatusBadRequest, "bad request")

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if body["error"] != "bad request" {
		t.Fatalf("expected error 'bad request', got %q", body["error"])
	}
	if len(body) != 1 {
		t.Fatalf("error object must contain exactly the error field, got %v", body)
	}
}

func TestStorePutGet(t *testing.T) {
	s := NewStore()
	f := Flag{Key: "a", Enabled: true, Description: "desc", RolloutPercent: 50}

	if err := s.Put(f); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, ok := s.Get("a")
	if !ok {
		t.Fatalf("expected flag to exist")
	}
	if got != f {
		t.Fatalf("expected %+v, got %+v", f, got)
	}
}

func TestStorePutDuplicate(t *testing.T) {
	s := NewStore()
	if err := s.Put(Flag{Key: "a"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := s.Put(Flag{Key: "a"}); err != ErrKeyExists {
		t.Fatalf("expected ErrKeyExists, got %v", err)
	}
}

func TestStoreListSorted(t *testing.T) {
	s := NewStore()
	for _, k := range []string{"c", "a", "b"} {
		if err := s.Put(Flag{Key: k}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	list := s.List()
	if len(list) != 3 {
		t.Fatalf("expected 3 flags, got %d", len(list))
	}
	for i, want := range []string{"a", "b", "c"} {
		if list[i].Key != want {
			t.Fatalf("expected key %q at position %d, got %q", want, i, list[i].Key)
		}
	}
}

func TestStoreListEmpty(t *testing.T) {
	s := NewStore()
	list := s.List()
	if len(list) != 0 {
		t.Fatalf("expected empty list, got %d entries", len(list))
	}
}

func TestStoreUpdate(t *testing.T) {
	s := NewStore()
	if err := s.Put(Flag{Key: "a", Enabled: false}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, ok := s.Update("a", Flag{Enabled: true, Description: "updated", RolloutPercent: 100})
	if !ok {
		t.Fatalf("expected update to succeed")
	}
	if got.Key != "a" || !got.Enabled || got.Description != "updated" || got.RolloutPercent != 100 {
		t.Fatalf("unexpected updated flag: %+v", got)
	}

	if _, ok := s.Update("missing", Flag{Enabled: true}); ok {
		t.Fatalf("expected update of missing key to fail")
	}
}

func TestStoreDelete(t *testing.T) {
	s := NewStore()
	if err := s.Put(Flag{Key: "a"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !s.Delete("a") {
		t.Fatalf("expected delete to succeed")
	}
	if _, ok := s.Get("a"); ok {
		t.Fatalf("expected flag to be gone")
	}
	if s.Delete("a") {
		t.Fatalf("expected delete of missing key to fail")
	}
}
