package main

import (
	"encoding/json"
	"fmt"
	"github/anansi-1/Chirpy/internal/database"
	"github/anansi-1/Chirpy/internal/auth"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func handleHealthzfunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) handleMetrics(w http.ResponseWriter, r *http.Request) {
	x := cfg.getFileServerHits()
	html := `<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(html, x)))
}

func (cfg *apiConfig) handleReset(w http.ResponseWriter, r *http.Request) {

	if cfg.platform != "dev" {
		respondWithError(w, http.StatusForbidden, "Your are not authorized to acces this")
		return
	}

	// cfg.resetFileServerHits()
	// w.WriteHeader(http.StatusOK)

	err := cfg.dbQueries.DeleteAllUsers(r.Context())
	if err != nil {
		http.Error(w, "Failed to delete users", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

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

func (cfg *apiConfig) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	type createUserRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type UserResponse struct {
		ID        string `json:"id"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		Email     string `json:"email"`
	}

	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if req.Email == "" || req.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Email and password are required")
		return
	}
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error hashing password")
		return
	}
	user, err := cfg.dbQueries.CreateUser(r.Context(), database.CreateUserParams{
		Email:          req.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error creating user")
		return
	}
	resp := UserResponse{
		ID:        user.ID.String(),
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
		Email:     user.Email,
	}

	respondWithJSON(w, http.StatusCreated, resp)
}


func (cfg *apiConfig) handleCreateChirp(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	type ChirpRequest struct {
		Body   string `json:"body"`
		UserID string `json:"user_id"`
	}

	type ChirpResponse struct {
		ID        string `json:"id"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		Body      string `json:"body"`
		UserID    string `json:"user_id"`
	}

	var newChirp ChirpRequest

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&newChirp); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	userUUID, err := uuid.Parse(newChirp.UserID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user_id format")
		return
	}

	chirp, err := cfg.dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   newChirp.Body,
		UserID: uuid.NullUUID{UUID: userUUID, Valid: true},
	})

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error Creating Chirp")
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

	chirp_rows,err := cfg.dbQueries.GetAllChirps(r.Context())
		if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error Getting Chirp")
		return
	}
	
	var chirps []ChirpResponse

	for _,chirp := range chirp_rows {

		resp := ChirpResponse{
		ID:        chirp.ID.String(),
		CreatedAt: chirp.CreatedAt.Format(time.RFC3339),
		UpdatedAt: chirp.UpdatedAt.Format(time.RFC3339),
		Body:      chirp.Body,
		UserID:    chirp.UserID.UUID.String(), 
	}
	chirps = append(chirps, resp)

}

	respondWithJSON(w,http.StatusOK,chirps)

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

