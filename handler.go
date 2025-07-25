package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
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

	if cfg.platform != "dev"{
		respondWithError(w,http.StatusForbidden,"Your are not authorized to acces this")
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

	type UserResponse struct {
	ID        string `json:"id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Email     string `json:"email"`
}

	defer r.Body.Close()
	var newUser User

	Decoder := json.NewDecoder(r.Body)
	if err := Decoder.Decode(&newUser); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	user, err := cfg.dbQueries.CreateUser(r.Context(), newUser.Email)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error Creating User")
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
