package main

import (
	"hash/fnv"
	"net/http"
)

func stableHash(key, user string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(key))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write([]byte(user))
	return h.Sum64()
}

const maxUserLength = 256

func (a *API) handleEvaluate(w http.ResponseWriter, r *http.Request) {
	user := r.URL.Query().Get("user")
	if user == "" {
		writeError(w, http.StatusBadRequest, "user is required")
		return
	}
	if len(user) > maxUserLength {
		writeError(w, http.StatusBadRequest, "user must be at most 256 characters")
		return
	}

	key := r.PathValue("key")
	flag, ok := a.store.Get(key)
	if !ok {
		writeError(w, http.StatusNotFound, "flag not found")
		return
	}

	enabled := false
	switch {
	case !flag.Enabled:
		enabled = false
	case flag.RolloutPercent >= 100:
		enabled = true
	case flag.RolloutPercent <= 0:
		enabled = false
	default:
		enabled = stableHash(flag.Key, user)%100 < uint64(flag.RolloutPercent)
	}

	writeJSON(w, http.StatusOK, map[string]bool{"enabled": enabled})
}
