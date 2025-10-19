package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type envelope map[string]interface{}

func (cfg *apiConfig) errorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	resp, err := json.Marshal(envelope{"error": message})
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		return
	}
	w.Write(resp)
}

func (cfg *apiConfig) readJSON(r *http.Request, dst interface{}) error {
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(dst)
	if err != nil {
		return fmt.Errorf("error decoding json")
	}
	return nil
}

func (cfg *apiConfig) writeJSON(w http.ResponseWriter, status int, data interface{}, headers http.Header) error {
	jsonRes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(jsonRes)
	return nil
}
