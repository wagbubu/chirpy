package main

import "net/http"

func (api *apiConfig) routes() http.Handler {
	fileServer := http.StripPrefix("/app", http.FileServer(http.Dir("./internal/static/")))
	mux := http.NewServeMux()

	mux.Handle("/app/", api.middlewareMetricsInc(fileServer))
	mux.HandleFunc("GET /api/healthz", api.healthCheck)
	mux.HandleFunc("GET /admin/metrics", api.metrics)
	mux.HandleFunc("POST /admin/reset", api.reset)
	mux.HandleFunc("POST /api/users", api.createUser)
	mux.HandleFunc("POST /api/login", api.loginUser)
	mux.HandleFunc("POST /api/refresh", api.refresh)
	mux.HandleFunc("POST /api/revoke", api.revoke)
	mux.HandleFunc("POST /api/chirps", api.authenticate(api.createChirp))
	mux.HandleFunc("GET /api/chirps", api.getAllChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", api.getChirp)

	return mux
}
