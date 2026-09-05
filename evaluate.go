package main

import "net/http"

func stableHash(key, user string) uint64 {
	return 0
}

func (a *API) handleEvaluate(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "not implemented")
}
