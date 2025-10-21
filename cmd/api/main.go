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
	jwtSecret      string
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	jwtSecret := os.Getenv("JWT_SECRET")
	platform := os.Getenv("PLATFORM")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}

	api := &apiConfig{
		db:             database.New(db),
		fileserverHits: atomic.Int32{},
		platform:       platform,
		jwtSecret:      jwtSecret,
	}

	srv := http.Server{
		Handler: api.routes(),
		Addr:    ":8080",
	}

	fmt.Println("Starting Server...")
	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
