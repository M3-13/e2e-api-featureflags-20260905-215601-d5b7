package main

import (
	"log"
	"net/http"
	"os"
)

type API struct {
	store *Store
}

func NewAPI() *API {
	return &API{store: NewStore()}
}

func (a *API) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", route(map[string]http.HandlerFunc{
		http.MethodGet: a.handleHealthz,
	}))

	mux.HandleFunc("/flags", route(map[string]http.HandlerFunc{
		http.MethodPost: a.handleCreateFlag,
		http.MethodGet:  a.handleListFlags,
	}))

	mux.HandleFunc("/flags/{key}", route(map[string]http.HandlerFunc{
		http.MethodGet:    a.handleGetFlag,
		http.MethodPut:    a.handleUpdateFlag,
		http.MethodDelete: a.handleDeleteFlag,
	}))

	mux.HandleFunc("/flags/{key}/evaluate", route(map[string]http.HandlerFunc{
		http.MethodGet: a.handleEvaluate,
	}))

	mux.HandleFunc("/", notFound)

	return withLogging(mux)
}

func (a *API) handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func main() {
	api := NewAPI()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: api.routes(),
	}

	log.Printf("featureflags service listening on :%s", port)
	log.Fatal(server.ListenAndServe())
}
