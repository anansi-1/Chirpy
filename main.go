package main

import (
	"database/sql"
	"github/anansi-1/Chirpy/internal/database"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
	platform       string
	tokenSecret    string
	apiKey         string
}
type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
}

func main() {
	const port = "8080"
	const filepathRoot = "."

	godotenv.Load()

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}
	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatal("PLATFORM must be set")
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is not set")
	}
	polkaKey := os.Getenv("POLKA_KEY")
	if polkaKey == "" {
		log.Fatal("POLKA_KEY environment variable is not set")
	}

	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error connecting to database: %s", err)
	}
	dbQueries := database.New(dbConn)

	apiConfig := apiConfig{
		dbQueries:   dbQueries,
		platform:    os.Getenv("PLATFORM"),
		tokenSecret: os.Getenv("JWT_SECRET"),
		apiKey:      os.Getenv("POLKA_KEY"),
	}

	mux := http.NewServeMux()
	fsHandler := apiConfig.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot))))


	mux.Handle("/app/", fsHandler)
	mux.HandleFunc("GET /api/healthz", handleHealthzfunc)

	mux.HandleFunc("POST /api/polka/webhooks", apiConfig.handleUpgradeWebhook)

	mux.HandleFunc("POST /api/login", apiConfig.handleLogin)
	mux.HandleFunc("POST /api/refresh", apiConfig.handleRefreshAccessToken)
	mux.HandleFunc("POST /api/revoke", apiConfig.handleRevokeRefreshToken)

	mux.HandleFunc("POST /api/users", apiConfig.handleCreateUser)
	mux.HandleFunc("PUT /api/users", apiConfig.handleUpdateUser)
	
	mux.HandleFunc("GET /api/chirps", apiConfig.handleGetChirps)
	mux.HandleFunc("POST /api/chirps", apiConfig.handleCreateChirp)
	mux.HandleFunc("POST /api/validate_chirp", handleValidateChirp)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiConfig.handleGetChirpsByID)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiConfig.handleDeleteChirp)
	
	mux.HandleFunc("GET /admin/metrics", apiConfig.handleMetrics)
	mux.HandleFunc("POST /admin/reset", apiConfig.handleReset)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(srv.ListenAndServe())

}
