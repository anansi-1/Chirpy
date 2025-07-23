package main

import (
	"encoding/json"
	"net/http"
)

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) error {

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(data)
	return nil
}

func respondWithError(w http.ResponseWriter, code int, msg string) error {

	return respondWithJSON(w, code, map[string]string{
		"error": msg})
}
