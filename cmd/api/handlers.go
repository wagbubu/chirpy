package main

import (
	"chirpy/internal/database"
	"chirpy/internal/dto"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

func (api *apiConfig) createChirp(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	err := api.readJSON(r, &input)
	if err != nil {
		api.errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if len(input.Body) > 140 {
		api.errorResponse(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	splittedWords := strings.Fields(input.Body)

	for i := range splittedWords {
		if _, ok := profane[strings.ToLower(splittedWords[i])]; ok {
			splittedWords[i] = "****"
		}
	}

	modifiedResp := strings.Join(splittedWords, " ")
	chirp := database.CreateChirpParams{Body: modifiedResp, UserID: input.UserID}

	newChirp, err := api.db.CreateChirp(r.Context(), chirp)
	if err != nil {
		api.errorResponse(w, http.StatusInternalServerError, "failed to create chirp in the database")
	}

	err = api.writeJSON(w, http.StatusCreated, dto.Chirp{ID: newChirp.ID, CreatedAt: newChirp.CreatedAt, UpdatedAt: newChirp.UpdatedAt, Body: newChirp.Body, UserID: newChirp.UserID}, nil)
	if err != nil {
		api.errorResponse(w, http.StatusInternalServerError, err.Error())
	}

}

func (api *apiConfig) createUser(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email string `json:"email"`
	}

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		api.errorResponse(w, http.StatusBadRequest, "error decoding request body")
		return
	}
	if input.Email == "" {
		api.errorResponse(w, http.StatusBadRequest, "email cannot be empty")
		return
	}

	user, err := api.db.CreateUser(r.Context(), input.Email)
	if err != nil {
		fmt.Println(err)
		api.errorResponse(w, http.StatusBadRequest, "error creating user")
		return
	}

	err = api.writeJSON(w, http.StatusCreated, &dto.User{ID: user.ID, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt, Email: user.Email}, nil)
	if err != nil {
		api.errorResponse(w, http.StatusInternalServerError, "something went wrong")
		return
	}

}

func (api *apiConfig) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (api *apiConfig) metrics(w http.ResponseWriter, r *http.Request) {
	hits := fmt.Sprintf(`
	<html>
		<body>
			<h1>Welcome, Chirpy Admin</h1>
			<p>Chirpy has been visited %d times!</p>
		</body>
	</html>`, api.fileserverHits.Load())

	w.Header().Set("Content-Type", "text/html")

	w.Write([]byte(hits))
}

func (api *apiConfig) reset(w http.ResponseWriter, r *http.Request) {
	if api.platform != "dev" {
		api.errorResponse(w, http.StatusForbidden, "Forbidden")
		return
	}

	if err := api.db.ResetUsers(r.Context()); err != nil {
		api.errorResponse(w, http.StatusInternalServerError, "failed to reset users")
		return
	}

	api.writeJSON(w, http.StatusOK, envelope{"message": "users table have been reset"}, nil)
}
