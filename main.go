package main

import (
	"log"
	"net/http"
	"os"
	"time"
)

var version = "dev"

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

	return withLogging(withAuth(mux))
}

func (a *API) handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "version": version})
}

func main() {
	api := NewAPI()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	bindAddr := os.Getenv("BIND_ADDR")
	if bindAddr == "" {
		bindAddr = "127.0.0.1"
	}

	server := &http.Server{
		Addr:              bindAddr + ":" + port,
		Handler:           api.routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("featureflags service listening on %s", server.Addr)

	cert := os.Getenv("TLS_CERT")
	key := os.Getenv("TLS_KEY")
	if cert != "" && key != "" {
		log.Fatal(server.ListenAndServeTLS(cert, key))
	}
	log.Fatal(server.ListenAndServe())
}
