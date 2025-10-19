package main

import (
	"chirpy/internal/database"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	db             *database.Queries
	fileserverHits atomic.Int32
	platform       string
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}

	api := &apiConfig{
		db:             database.New(db),
		fileserverHits: atomic.Int32{},
		platform:       platform,
	}

	fileServer := http.StripPrefix("/app", http.FileServer(http.Dir("./internal/static/")))
	mux := http.NewServeMux()

	mux.Handle("/app/", api.middlewareMetricsInc(fileServer))
	mux.HandleFunc("GET /api/healthz", api.healthCheck)
	mux.HandleFunc("GET /admin/metrics", api.metrics)
	mux.HandleFunc("POST /admin/reset", api.reset)
	mux.HandleFunc("POST /api/users", api.createUser)
	mux.HandleFunc("POST /api/chirps", api.createChirp)

	srv := http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	fmt.Println("Starting Server...")
	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
