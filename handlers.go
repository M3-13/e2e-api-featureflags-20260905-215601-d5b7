package main

import (
	"encoding/json"
	"errors"
	"net/http"
)

const maxBodyBytes = 1 << 20 // 1 MiB

type createFlagRequest struct {
	Key            string `json:"key"`
	Enabled        *bool  `json:"enabled"`
	Description    string `json:"description"`
	RolloutPercent int    `json:"rollout_percent"`
}

type updateFlagRequest struct {
	Enabled        *bool  `json:"enabled"`
	Description    string `json:"description"`
	RolloutPercent int    `json:"rollout_percent"`
}

func decodeBody(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(dst); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
			return false
		}
		writeError(w, http.StatusBadRequest, "invalid request body")
		return false
	}
	return true
}

func (a *API) handleCreateFlag(w http.ResponseWriter, r *http.Request) {
	var req createFlagRequest
	if !decodeBody(w, r, &req) {
		return
	}

	if req.Key == "" {
		writeError(w, http.StatusBadRequest, "key is required")
		return
	}
	if req.Enabled == nil {
		writeError(w, http.StatusBadRequest, "enabled is required")
		return
	}
	if req.RolloutPercent < 0 || req.RolloutPercent > 100 {
		writeError(w, http.StatusBadRequest, "rollout_percent must be between 0 and 100")
		return
	}

	flag := Flag{
		Key:            req.Key,
		Enabled:        *req.Enabled,
		Description:    req.Description,
		RolloutPercent: req.RolloutPercent,
	}
	if err := a.store.Put(flag); err != nil {
		if errors.Is(err, ErrKeyExists) {
			writeError(w, http.StatusConflict, "key already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusCreated, flag)
}

func (a *API) handleListFlags(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.store.List())
}

func (a *API) handleGetFlag(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	flag, ok := a.store.Get(key)
	if !ok {
		writeError(w, http.StatusNotFound, "flag not found")
		return
	}
	writeJSON(w, http.StatusOK, flag)
}

func (a *API) handleUpdateFlag(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	var req updateFlagRequest
	if !decodeBody(w, r, &req) {
		return
	}

	if req.Enabled == nil {
		writeError(w, http.StatusBadRequest, "enabled is required")
		return
	}
	if req.RolloutPercent < 0 || req.RolloutPercent > 100 {
		writeError(w, http.StatusBadRequest, "rollout_percent must be between 0 and 100")
		return
	}

	flag := Flag{
		Enabled:        *req.Enabled,
		Description:    req.Description,
		RolloutPercent: req.RolloutPercent,
	}
	updated, ok := a.store.Update(key, flag)
	if !ok {
		writeError(w, http.StatusNotFound, "flag not found")
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (a *API) handleDeleteFlag(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if !a.store.Delete(key) {
		writeError(w, http.StatusNotFound, "flag not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
