package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (cfg *apiConfig) Router() http.Handler {
	r := chi.NewRouter()

	apiRouter := chi.NewRouter()
	apiAdmin := chi.NewRouter()

	r.Handle("/app*", cfg.middlewaremetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))

	apiRouter.Post("/validate_chirp", handlerChirpsValidate)
	apiRouter.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	apiAdmin.Post("/reset", cfg.handleReset)
	apiAdmin.Get("/metrics", cfg.handleGetMetrics)

	r.Mount("/api", apiRouter)
	r.Mount("/admin", apiAdmin)

	return r
}
