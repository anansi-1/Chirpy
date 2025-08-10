package main

import (
	"github/anansi-1/Chirpy/internal/auth"
	"net/http"
	"time"

)

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