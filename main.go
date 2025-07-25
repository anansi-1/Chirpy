package main

import (
	"database/sql"
	"github/anansi-1/Chirpy/internal/database"
	"log"
	"net/http"
	"os"
	"time"
	"sync/atomic"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
	platform 		string
}
type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func main() {

	godotenv.Load()

	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	dbQueries := database.New(db)

	apiConfig := apiConfig{
		dbQueries: dbQueries,
		platform: os.Getenv("PLATFORM"),
	}

	const port = "8080"
	const filepathRoot = "."

	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir(filepathRoot))

	mux.Handle("/app/", apiConfig.middlewareMetricsInc(http.StripPrefix("/app", fs)))

	mux.HandleFunc("GET /api/healthz", handleHealthzfunc)
	mux.HandleFunc("GET /admin/metrics", apiConfig.handleMetrics)
	mux.HandleFunc("POST /admin/reset", apiConfig.handleReset)
	mux.HandleFunc("POST /api/validate_chirp", handleValidateChirp)
	mux.HandleFunc("POST /api/users",apiConfig.handleCreateUser)

	srv := &http.Server{ // a struct that describes the server configuration
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(srv.ListenAndServe())

}
