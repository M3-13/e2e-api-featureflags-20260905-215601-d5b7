package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
)

func itoa(n int) string {
	return strconv.Itoa(n)
}

func doRequest(t *testing.T, api *API, method, path string, body []byte) *httptest.ResponseRecorder {
	t.Helper()
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, path, bytes.NewReader(body))
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	rr := httptest.NewRecorder()
	api.routes().ServeHTTP(rr, req)
	return rr
}

func decodeFlag(t *testing.T, rr *httptest.ResponseRecorder) Flag {
	t.Helper()
	var f Flag
	if err := json.Unmarshal(rr.Body.Bytes(), &f); err != nil {
		t.Fatalf("invalid flag json: %v", err)
	}
	return f
}

func decodeFlagList(t *testing.T, rr *httptest.ResponseRecorder) []Flag {
	t.Helper()
	var list []Flag
	if err := json.Unmarshal(rr.Body.Bytes(), &list); err != nil {
		t.Fatalf("invalid list json: %v", err)
	}
	return list
}

func TestCreateFlagSuccess(t *testing.T) {
	api := NewAPI()
	body := []byte(`{"key":"feature-a","enabled":true,"description":"hello","rollout_percent":50}`)

	rr := doRequest(t, api, http.MethodPost, "/flags", body)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	f := decodeFlag(t, rr)
	if f.Key != "feature-a" || !f.Enabled || f.Description != "hello" || f.RolloutPercent != 50 {
		t.Fatalf("unexpected flag: %+v", f)
	}
}

func TestCreateFlagDefaults(t *testing.T) {
	api := NewAPI()
	body := []byte(`{"key":"feature-b","enabled":false}`)

	rr := doRequest(t, api, http.MethodPost, "/flags", body)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	f := decodeFlag(t, rr)
	if f.Key != "feature-b" || f.Enabled || f.Description != "" || f.RolloutPercent != 0 {
		t.Fatalf("unexpected flag: %+v", f)
	}
}

func TestCreateFlagMissingKey(t *testing.T) {
	api := NewAPI()
	rr := doRequest(t, api, http.MethodPost, "/flags", []byte(`{"enabled":true}`))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
	assertErrorJSON(t, rr)
}

func TestCreateFlagEmptyKey(t *testing.T) {
	api := NewAPI()
	rr := doRequest(t, api, http.MethodPost, "/flags", []byte(`{"key":"","enabled":true}`))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
	assertErrorJSON(t, rr)
}

func TestCreateFlagMissingEnabled(t *testing.T) {
	api := NewAPI()
	rr := doRequest(t, api, http.MethodPost, "/flags", []byte(`{"key":"feature-c"}`))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
	assertErrorJSON(t, rr)
}

func TestCreateFlagRolloutOutOfRange(t *testing.T) {
	api := NewAPI()
	for _, v := range []string{"-1", "101"} {
		body := []byte(`{"key":"feature-d","enabled":true,"rollout_percent":` + v + `}`)
		rr := doRequest(t, api, http.MethodPost, "/flags", body)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for rollout_percent %s, got %d", v, rr.Code)
		}
		assertErrorJSON(t, rr)
	}
}

func TestCreateFlagDuplicate(t *testing.T) {
	api := NewAPI()
	body := []byte(`{"key":"dup","enabled":true}`)

	if rr := doRequest(t, api, http.MethodPost, "/flags", body); rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}
	rr := doRequest(t, api, http.MethodPost, "/flags", body)
	if rr.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rr.Code)
	}
	assertErrorJSON(t, rr)
}

func TestCreateFlagInvalidJSON(t *testing.T) {
	api := NewAPI()
	rr := doRequest(t, api, http.MethodPost, "/flags", []byte(`{not json`))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
	assertErrorJSON(t, rr)
}

func TestCreateFlagTooLarge(t *testing.T) {
	api := NewAPI()
	big := []byte(`{"key":"` + strings.Repeat("a", maxBodyBytes) + `","enabled":true}`)
	rr := doRequest(t, api, http.MethodPost, "/flags", big)
	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", rr.Code)
	}
	assertErrorJSON(t, rr)
}

func TestListFlagsEmpty(t *testing.T) {
	api := NewAPI()
	rr := doRequest(t, api, http.MethodGet, "/flags", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	list := decodeFlagList(t, rr)
	if len(list) != 0 {
		t.Fatalf("expected empty list, got %d entries", len(list))
	}
}

func TestListFlagsSorted(t *testing.T) {
	api := NewAPI()
	for _, k := range []string{"c", "a", "b"} {
		body := []byte(`{"key":"` + k + `","enabled":true}`)
		if rr := doRequest(t, api, http.MethodPost, "/flags", body); rr.Code != http.StatusCreated {
			t.Fatalf("expected 201 for key %q, got %d", k, rr.Code)
		}
	}

	rr := doRequest(t, api, http.MethodGet, "/flags", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	list := decodeFlagList(t, rr)
	if len(list) != 3 {
		t.Fatalf("expected 3 flags, got %d", len(list))
	}
	for i, want := range []string{"a", "b", "c"} {
		if list[i].Key != want {
			t.Fatalf("expected key %q at position %d, got %q", want, i, list[i].Key)
		}
	}
}

func TestGetFlag(t *testing.T) {
	api := NewAPI()
	if rr := doRequest(t, api, http.MethodPost, "/flags", []byte(`{"key":"get","enabled":true}`)); rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}

	rr := doRequest(t, api, http.MethodGet, "/flags/get", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	f := decodeFlag(t, rr)
	if f.Key != "get" || !f.Enabled {
		t.Fatalf("unexpected flag: %+v", f)
	}
}

func TestGetFlagNotFound(t *testing.T) {
	api := NewAPI()
	rr := doRequest(t, api, http.MethodGet, "/flags/missing", nil)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
	assertErrorJSON(t, rr)
}

func TestUpdateFlag(t *testing.T) {
	api := NewAPI()
	if rr := doRequest(t, api, http.MethodPost, "/flags", []byte(`{"key":"upd","enabled":false}`)); rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}

	rr := doRequest(t, api, http.MethodPut, "/flags/upd", []byte(`{"enabled":true,"description":"changed","rollout_percent":75}`))
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	f := decodeFlag(t, rr)
	if f.Key != "upd" || !f.Enabled || f.Description != "changed" || f.RolloutPercent != 75 {
		t.Fatalf("unexpected flag: %+v", f)
	}
}

func TestUpdateFlagMissingEnabled(t *testing.T) {
	api := NewAPI()
	if rr := doRequest(t, api, http.MethodPost, "/flags", []byte(`{"key":"upd2","enabled":true}`)); rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}

	rr := doRequest(t, api, http.MethodPut, "/flags/upd2", []byte(`{"description":"no enabled"}`))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
	assertErrorJSON(t, rr)
}

func TestUpdateFlagRolloutOutOfRange(t *testing.T) {
	api := NewAPI()
	if rr := doRequest(t, api, http.MethodPost, "/flags", []byte(`{"key":"upd3","enabled":true}`)); rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}

	rr := doRequest(t, api, http.MethodPut, "/flags/upd3", []byte(`{"enabled":true,"rollout_percent":101}`))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
	assertErrorJSON(t, rr)
}

func TestUpdateFlagNotFound(t *testing.T) {
	api := NewAPI()
	rr := doRequest(t, api, http.MethodPut, "/flags/missing", []byte(`{"enabled":true}`))
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
	assertErrorJSON(t, rr)
}

func TestUpdateFlagTooLarge(t *testing.T) {
	api := NewAPI()
	if rr := doRequest(t, api, http.MethodPost, "/flags", []byte(`{"key":"upd4","enabled":true}`)); rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}

	big := []byte(`{"enabled":true,"description":"` + strings.Repeat("a", maxBodyBytes) + `"}`)
	rr := doRequest(t, api, http.MethodPut, "/flags/upd4", big)
	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", rr.Code)
	}
	assertErrorJSON(t, rr)
}

func TestDeleteFlag(t *testing.T) {
	api := NewAPI()
	if rr := doRequest(t, api, http.MethodPost, "/flags", []byte(`{"key":"del","enabled":true}`)); rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}

	rr := doRequest(t, api, http.MethodDelete, "/flags/del", nil)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}

	if rr := doRequest(t, api, http.MethodGet, "/flags/del", nil); rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d", rr.Code)
	}
}

func TestDeleteFlagNotFound(t *testing.T) {
	api := NewAPI()
	rr := doRequest(t, api, http.MethodDelete, "/flags/missing", nil)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
	assertErrorJSON(t, rr)
}

func TestConcurrentAccess(t *testing.T) {
	api := NewAPI()

	const workers = 50
	const perWorker = 20

	var wg sync.WaitGroup
	wg.Add(workers)
	for w := 0; w < workers; w++ {
		go func(id int) {
			defer wg.Done()
			for i := 0; i < perWorker; i++ {
				key := "flag-" + itoa(id) + "-" + itoa(i)
				body := []byte(`{"key":"` + key + `","enabled":true}`)
				if rr := doRequest(t, api, http.MethodPost, "/flags", body); rr.Code != http.StatusCreated {
					t.Errorf("expected 201 for %q, got %d", key, rr.Code)
					return
				}
				if rr := doRequest(t, api, http.MethodGet, "/flags/"+key, nil); rr.Code != http.StatusOK {
					t.Errorf("expected 200 for %q, got %d", key, rr.Code)
					return
				}
			}
		}(w)
	}
	wg.Wait()

	rr := doRequest(t, api, http.MethodGet, "/flags", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	list := decodeFlagList(t, rr)
	if len(list) != workers*perWorker {
		t.Fatalf("expected %d flags, got %d", workers*perWorker, len(list))
	}
}
