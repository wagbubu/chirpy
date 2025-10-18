package main

import (
	"fmt"
	"net/http"
)

func (cfg *apiConfig) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	w.Write([]byte("OK"))
}

func (cfg *apiConfig) metrics(w http.ResponseWriter, r *http.Request) {
	hits := fmt.Sprintf(`
	<html>
		<body>
			<h1>Welcome, Chirpy Admin</h1>
			<p>Chirpy has been visited %d times!</p>
		</body>
	</html>`, cfg.fileserverHits.Load())

	w.Header().Set("Content-Type", "text/html")

	w.Write([]byte(hits))
}

func (cfg *apiConfig) resetMetrics(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	hits := fmt.Sprintf("Hits: %d", cfg.fileserverHits.Load())
	w.Write([]byte(hits))
}

func (cfg *apiConfig) validateChirp(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Body string `json:"body"`
	}
	err := cfg.readJSON(r, &input)
	if err != nil {
		cfg.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	if len(input.Body) > 140 {
		cfg.errorResponse(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	err = cfg.writeJSON(w, envelope{"valid": true})
	if err != nil {
		cfg.errorResponse(w, http.StatusInternalServerError, err.Error())
	}
}
