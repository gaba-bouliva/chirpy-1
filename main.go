package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewaremetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)

	})
}

func main() {
	port := 8080
	apiCfg := apiConfig{}

	srv := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: apiCfg.Router(),
	}

	log.Printf("Starting server on :%d", port)
	log.Fatal(srv.ListenAndServe())
}
