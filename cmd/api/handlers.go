package main

import (
	"chirpy/internal/auth"
	"chirpy/internal/database"
	"chirpy/internal/dto"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (api *apiConfig) getChirp(w http.ResponseWriter, r *http.Request) {
	chirpID := r.PathValue("chirpID")
	if chirpID == "" {
		api.errorResponse(w, http.StatusNotFound, "chirp not found")
		return
	}

	id, err := uuid.Parse(chirpID)
	if err != nil {
		api.errorResponse(w, http.StatusNotFound, "chirp not found")
		return
	}

	chirp, err := api.db.GetChirp(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			api.errorResponse(w, http.StatusNotFound, "chirp not found")
		default:
			api.errorResponse(w, http.StatusInternalServerError, "something went wrong")
		}
		return
	}

	err = api.writeJSON(w, http.StatusOK, dto.Chirp{ID: chirp.ID, Body: chirp.Body, CreatedAt: chirp.CreatedAt, UpdatedAt: chirp.UpdatedAt, UserID: chirp.UserID}, nil)
	if err != nil {
		api.errorResponse(w, http.StatusInternalServerError, "error writing a response")
		return
	}
}

func (api *apiConfig) getAllChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := api.db.GetAllChirps(r.Context())
	if err != nil {
		api.errorResponse(w, http.StatusInternalServerError, "failed to get all chirps")
		return
	}

	chirpsResponse := make([]dto.Chirp, 0, len(chirps))
	for _, chirp := range chirps {
		chirpsResponse = append(chirpsResponse, dto.Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		})
	}

	err = api.writeJSON(w, http.StatusOK, chirpsResponse, nil)
	if err != nil {
		api.errorResponse(w, http.StatusInternalServerError, "error writing a response")
		return
	}
}

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
	chirp := database.CreateChirpParams{Body: modifiedResp, UserID: api.contextGetUserID(r)}

	newChirp, err := api.db.CreateChirp(r.Context(), chirp)
	if err != nil {
		api.errorResponse(w, http.StatusInternalServerError, "failed to create chirp in the database")
		return
	}

	err = api.writeJSON(w, http.StatusCreated, dto.Chirp{ID: newChirp.ID, CreatedAt: newChirp.CreatedAt, UpdatedAt: newChirp.UpdatedAt, Body: newChirp.Body, UserID: newChirp.UserID}, nil)
	if err != nil {
		api.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
}

func (api *apiConfig) loginUser(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := api.readJSON(r, &input)
	if err != nil {
		api.errorResponse(w, http.StatusBadRequest, "malformed form request")
		return
	}

	user, err := api.db.GetUser(r.Context(), input.Email)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			api.errorResponse(w, http.StatusUnauthorized, "invalid login")
		default:
			api.errorResponse(w, http.StatusInternalServerError, "something went wrong")
		}
		return
	}

	match, err := auth.CheckPasswordHash(input.Password, user.HashedPassword)
	if err != nil {
		api.errorResponse(w, http.StatusInternalServerError, "something went wrong")
		return
	}
	if !match {
		api.errorResponse(w, http.StatusUnauthorized, "invalid login credentials")
		return
	}

	token, err := auth.MakeJWT(user.ID, api.jwtSecret, time.Hour)
	if err != nil {
		api.errorResponse(w, http.StatusInternalServerError, "something went wrong!")
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		api.errorResponse(w, http.StatusInternalServerError, "something went wrong!")
		return
	}

	refreshTokenToStoreInDB := database.InsertRefreshTokenParams{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 60).UTC(),
	}

	err = api.db.InsertRefreshToken(r.Context(), refreshTokenToStoreInDB)
	if err != nil {
		api.errorResponse(w, http.StatusInternalServerError, "something went wrong!")
		return
	}

	resp := dto.User{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        token,
		RefreshToken: refreshToken,
	}

	r.Context()
	err = api.writeJSON(w, http.StatusOK, resp, nil)
	if err != nil {
		api.errorResponse(w, http.StatusInternalServerError, "something went wrong!")
		return
	}
}

func (api *apiConfig) refresh(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrInvalidToken):
			api.errorResponse(w, http.StatusUnauthorized, "unauthorized")
		default:
			api.errorResponse(w, http.StatusInternalServerError, "something went wrong!")
		}
		return
	}

	userID, err := api.db.GetUserRefreshToken(r.Context(), refreshToken)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			api.errorResponse(w, http.StatusUnauthorized, "unauthorized")
		default:
			api.errorResponse(w, http.StatusInternalServerError, "something went wrong!")
		}
		return
	}
	fmt.Println("RAN UP TO HERE")
	token, err := auth.MakeJWT(userID, api.jwtSecret, time.Hour)
	if err != nil {

		api.errorResponse(w, http.StatusInternalServerError, "something went wrong!")
		return
	}

	err = api.writeJSON(w, http.StatusOK, envelope{"token": token}, nil)
	if err != nil {
		api.errorResponse(w, http.StatusInternalServerError, "something went wrong!")
		return
	}
}

func (api *apiConfig) revoke(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrInvalidToken):
			api.errorResponse(w, http.StatusUnauthorized, "unauthorized")
		default:
			api.errorResponse(w, http.StatusInternalServerError, "something went wrong!")
		}
		return
	}

	err = api.db.RevokeToken(r.Context(), token)
	if err != nil {
		api.errorResponse(w, http.StatusInternalServerError, "something went wrong!")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (api *apiConfig) createUser(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
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

	hashedPassword, err := auth.HashPassword(input.Password)
	if err != nil {
		log.Fatal("error hashing password")
		return
	}

	newUser := database.CreateUserParams{Email: input.Email, HashedPassword: hashedPassword}
	user, err := api.db.CreateUser(r.Context(), newUser)
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
