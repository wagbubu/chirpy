package main

import (
	"chirpy/internal/auth"
	"errors"
	"log"
	"net/http"
)

func (api *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		api.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (api *apiConfig) authenticate(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Authorization")

		token, err := auth.GetBearerToken(r.Header)
		if err != nil {
			api.errorResponse(w, http.StatusUnauthorized, err.Error())
		}
		userID, err := auth.ValidateJWT(token, api.jwtSecret)
		if err != nil {
			switch {
			case errors.Is(err, auth.ErrExpiredToken):
				api.errorResponse(w, http.StatusUnauthorized, err.Error())
			default:
				log.Fatal(err)
			}
			return
		}

		r = api.contextSetUserID(r, userID)

		next.ServeHTTP(w, r)
	})
}
