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

func main() {
	api := &apiConfig{
		fileserverHits: atomic.Int32{},
	}
	fileServer := http.StripPrefix("/app", http.FileServer(http.Dir("./internal/static/")))
	mux := http.NewServeMux()

	mux.Handle("/app/", api.middlewareMetricsInc(fileServer))
	mux.HandleFunc("GET /api/healthz", api.healthCheck)
	mux.HandleFunc("GET /admin/metrics", api.metrics)
	mux.HandleFunc("POST /admin/reset", api.resetMetrics)
	mux.HandleFunc("POST /api/validate_chirp", api.validateChirp)

	srv := http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	fmt.Println("Starting Server...")
	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
