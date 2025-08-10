package main

import (
	"encoding/json"
	"github/anansi-1/Chirpy/internal/auth"
	"github/anansi-1/Chirpy/internal/database"
	"net/http"
	"sort"
	"time"

	"github.com/google/uuid"
)

func handleValidateChirp(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var chirp struct {
		Body string `json:"body"`
	}

	if err := json.NewDecoder(r.Body).Decode(&chirp); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	if len(chirp.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	res := cleanedBody(chirp.Body)
	respondWithJSON(w, http.StatusOK, map[string]string{"cleaned_body": res})
}

func (cfg *apiConfig) handleCreateChirp(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	type ChirpRequest struct {
		Body string `json:"body"`
	}

	type ChirpResponse struct {
		ID        string `json:"id"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		Body      string `json:"body"`
		UserID    string `json:"user_id"`
	}

	tokenStr, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Missing or invalid Authorization header")
		return
	}

	userID, err := auth.ValidateJWT(tokenStr, cfg.tokenSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid or expired token")
		return
	}

	var newChirp ChirpRequest
	if err := json.NewDecoder(r.Body).Decode(&newChirp); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	chirp, err := cfg.dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   newChirp.Body,
		UserID: uuid.NullUUID{UUID: userID, Valid: true},
	})
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error creating chirp")
		return
	}

	resp := ChirpResponse{
		ID:        chirp.ID.String(),
		CreatedAt: chirp.CreatedAt.Format(time.RFC3339),
		UpdatedAt: chirp.UpdatedAt.Format(time.RFC3339),
		Body:      chirp.Body,
		UserID:    chirp.UserID.UUID.String(),
	}

	respondWithJSON(w, http.StatusCreated, resp)
}

func (cfg *apiConfig) handleGetChirps(w http.ResponseWriter, r *http.Request) {

	type ChirpResponse struct {
		ID        string `json:"id"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		Body      string `json:"body"`
		UserID    string `json:"user_id"`
	}

	s := r.URL.Query().Get("author_id")
	sortOrder := r.URL.Query().Get("sort")
	if sortOrder != "desc" {
		sortOrder = "asc"
	}

	var chirpRows []database.Chirp
	var err error

	if s != "" {
		authorID, err := uuid.Parse(s)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid UUID format")
			return
		}
		authorUUID := uuid.NullUUID{
			UUID:  authorID,
			Valid: true,
		}
		chirpRows, err = cfg.dbQueries.GetChirpsByAuthorID(r.Context(), authorUUID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error getting chirps by author")
			return
		}
	} else {
		chirpRows, err = cfg.dbQueries.GetAllChirps(r.Context())
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error getting chirps")
			return
		}
	}

	sort.Slice(chirpRows, func(i, j int) bool {
		if sortOrder == "asc" {
			return chirpRows[i].CreatedAt.Before(chirpRows[j].CreatedAt)
		}
		return chirpRows[i].CreatedAt.After(chirpRows[j].CreatedAt)
	})

	var chirps []ChirpResponse
	for _, chirp := range chirpRows {
		resp := ChirpResponse{
			ID:        chirp.ID.String(),
			CreatedAt: chirp.CreatedAt.Format(time.RFC3339),
			UpdatedAt: chirp.UpdatedAt.Format(time.RFC3339),
			Body:      chirp.Body,
			UserID:    chirp.UserID.UUID.String(),
		}
		chirps = append(chirps, resp)
	}

	respondWithJSON(w, http.StatusOK, chirps)
}

func (cfg *apiConfig) handleGetChirpsByID(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	type ChirpResponse struct {
		ID        string `json:"id"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		Body      string `json:"body"`
		UserID    string `json:"user_id"`
	}

	chirpID := r.PathValue("chirpID")
	chirpUUID, err := uuid.Parse(chirpID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID format")
		return
	}

	chirp, err := cfg.dbQueries.GetChirpsByID(r.Context(), chirpUUID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Chirp not found")
		return
	}

	resp := ChirpResponse{
		ID:        chirp.ID.String(),
		CreatedAt: chirp.CreatedAt.Format(time.RFC3339),
		UpdatedAt: chirp.UpdatedAt.Format(time.RFC3339),
		Body:      chirp.Body,
		UserID:    chirp.UserID.UUID.String(),
	}

	respondWithJSON(w, http.StatusOK, resp)
}

func (cfg *apiConfig) handleDeleteChirp(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	chirpID := r.PathValue("chirpID")
	chirpUUID, err := uuid.Parse(chirpID)
	if err != nil {
		http.Error(w, "Invalid chirp ID", http.StatusBadRequest)
		return
	}

	tokenStr, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Missing or invalid Authorization header")
		return
	}

	userID, err := auth.ValidateJWT(tokenStr, cfg.tokenSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid or expired token")
		return
	}

	chirp, err := cfg.dbQueries.GetChirpsByID(r.Context(), chirpUUID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Chirp not found")
		return
	}

	if !chirp.UserID.Valid || chirp.UserID.UUID.String() != userID.String() {
		respondWithError(w, http.StatusForbidden, "You are not the owner of this chirp")
		return
	}

	err = cfg.dbQueries.DeleteChirpsByID(r.Context(), chirpUUID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to delete chirp")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
