package main

import (
	"encoding/json"
	"fmt"
	"github/anansi-1/Chirpy/internal/auth"
	"github/anansi-1/Chirpy/internal/database"
	"net/http"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
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
		ID          string `json:"id"`
		CreatedAt   string `json:"created_at"`
		UpdatedAt   string `json:"updated_at"`
		Email       string `json:"email"`
		IsChirpyRed bool   `json:"is_chirpy_red"`
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
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			respondWithError(w, http.StatusConflict, "Email already exists")
			return
		}
		respondWithError(w, http.StatusBadRequest, "Error creating user")
		return
	}

	resp := UserResponse{
		ID:          user.ID.String(),
		CreatedAt:   user.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   user.UpdatedAt.Format(time.RFC3339),
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed.Bool,
	}

	respondWithJSON(w, http.StatusCreated, resp)
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

func (cfg *apiConfig) handleLogin(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	type UserLoginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type LoginResponse struct {
		ID           string `json:"id"`
		CreatedAt    string `json:"created_at"`
		UpdatedAt    string `json:"updated_at"`
		Email        string `json:"email"`
		IsChirpyRed  bool   `json:"is_chirpy_red"`
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}

	var req UserLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if req.Email == "" || req.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	user, err := cfg.dbQueries.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	err = auth.CheckPasswordHash(req.Password, user.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	accessToken, err := auth.MakeJWT(user.ID, cfg.tokenSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create JWT")
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create refresh token")
		return
	}

	refreshExpiresAt := time.Now().Add(60 * 24 * time.Hour)
	_, err = cfg.dbQueries.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: refreshExpiresAt,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to save refresh token")
		return
	}

	resp := LoginResponse{
		ID:           user.ID.String(),
		CreatedAt:    user.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    user.UpdatedAt.Format(time.RFC3339),
		Email:        user.Email,
		IsChirpyRed:  user.IsChirpyRed.Bool,
		Token:        accessToken,
		RefreshToken: refreshToken,
	}

	respondWithJSON(w, http.StatusOK, resp)
}

func (cfg *apiConfig) handleRefreshAccessToken(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Missing or invalid refresh token")
		return
	}

	tokenRecord, err := cfg.dbQueries.GetRefreshToken(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Refresh token not found")
		return
	}

	if tokenRecord.ExpiresAt.Before(time.Now()) || tokenRecord.RevokedAt.Valid {
		respondWithError(w, http.StatusUnauthorized, "Refresh token expired or revoked")
		return
	}

	newToken, err := auth.MakeJWT(tokenRecord.UserID, cfg.tokenSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create new token")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"token": newToken,
	})
}

func (cfg *apiConfig) handleRevokeRefreshToken(w http.ResponseWriter, r *http.Request) {

	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Missing or invalid refresh token")
		return
	}

	err = cfg.dbQueries.RevokeRefreshToken(r.Context(), refreshToken)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid or already revoked token")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
func (cfg *apiConfig) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

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

	type userUpdateRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type UserResponse struct {
		ID          string `json:"id"`
		CreatedAt   string `json:"created_at"`
		UpdatedAt   string `json:"updated_at"`
		Email       string `json:"email"`
		IsChirpyRed bool   `json:"is_chirpy_red"`
	}

	var userUpdate userUpdateRequest

	if err := json.NewDecoder(r.Body).Decode(&userUpdate); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if userUpdate.Email == "" || userUpdate.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	hashedPassword, err := auth.HashPassword(userUpdate.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error hashing password")
		return
	}

	user, err := cfg.dbQueries.UpdateUser(r.Context(), database.UpdateUserParams{
		ID:             userID,
		Email:          userUpdate.Email,
		HashedPassword: hashedPassword,
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to Update User info")
		return
	}

	resp := UserResponse{
		ID:          user.ID.String(),
		CreatedAt:   user.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   user.UpdatedAt.Format(time.RFC3339),
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed.Bool,
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

func (cfg *apiConfig) handleUpgradeWebhook(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Missing or invalid Authorization header")
		return
	}

	if apiKey != cfg.apiKey {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	type UpgradeRequest struct {
		Event string `json:"event"`
		Data  struct {
			UserID string `json:"user_id"`
		} `json:"data"`
	}

	var upgradeBody UpgradeRequest
	if err := json.NewDecoder(r.Body).Decode(&upgradeBody); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if upgradeBody.Event != "user.upgraded" {
		respondWithJSON(w, http.StatusNoContent, nil)
		return
	}

	userID, err := uuid.Parse(upgradeBody.Data.UserID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid UUID format")
		return
	}

	rowsAffected, err := cfg.dbQueries.UpgradeUser(r.Context(), userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to upgrade user")
		return
	}

	if rowsAffected == 0 {
		respondWithError(w, http.StatusNotFound, "User not found")
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}
