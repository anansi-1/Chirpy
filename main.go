package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github/anansi-1/Chirpy/internal/database"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries *database.Queries
}



func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)         
		next.ServeHTTP(w, r)               
	})
}

func (cfg *apiConfig) getFileServerHits() int32{
	return cfg.fileserverHits.Load()
}
func (cfg *apiConfig) resetFileServerHits() {
	 cfg.fileserverHits.Store(0)
}

func main() {

	godotenv.Load()

	dbURL := os.Getenv("DB_URL")
	db,err := sql.Open("postgres",dbURL)
	if err != nil {
		log.Fatal(err)
	}
	dbQueries := database.New(db)

	apiConfig := apiConfig{
		dbQueries: dbQueries,
	}
	const port = "8080"
	const filepathRoot = "."
	
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir(filepathRoot))

	// mux.Handle("/app/",http.StripPrefix("/app",fs))
	mux.Handle("/app/", apiConfig.middlewareMetricsInc(http.StripPrefix("/app",fs)))


	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type","text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("GET /admin/metrics",func(w http.ResponseWriter, r *http.Request) {
		x := apiConfig.getFileServerHits()
		text := fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`,x)
		w.Header().Set("Content-Type","text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(text))
	})

	
	mux.HandleFunc("POST /admin/reset",func(w http.ResponseWriter, r *http.Request) {
		apiConfig.resetFileServerHits()
	})

	mux.HandleFunc("POST /api/validate_chirp",func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		type msg struct{
			Body string `json:"body"`
		}

		type response struct{
			Cleaned_Body string `json:"cleaned_body"`
		}

		chirp := msg{}

		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&chirp); err != nil {
			respondWithError(w,http.StatusInternalServerError,"Something went wrong")
			return
		}

		if len(chirp.Body) > 140 {
			respondWithError(w,http.StatusBadRequest,"Chirp is too long")
			return
		}

		res := cleanedBody(chirp.Body)
		respondWithJSON(w,http.StatusOK,response{Cleaned_Body: res})


	})

	
	

	srv := &http.Server{  // a struct that describes the server configuration
		Addr: ":" + port,
		Handler: mux,
	}

	log.Printf("Serving on port: %s\n",port)
	log.Fatal(srv.ListenAndServe())

}

func respondWithJSON (w http.ResponseWriter, code int, payload interface{}) error{

	data,err := json.Marshal(payload)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(code)
	w.Write(data)
	return nil
}

func respondWithError (w http.ResponseWriter, code int, msg string) error{

	return respondWithJSON(w,code,map[string]string{
		"error":msg})
	}



func cleanedBody(s string) string{

	// wordsToBeRemoved := []string{"fornax","sharbert","kerfuffle"}
	var cleanedWords []string
	words := strings.Fields(s)
	for _,word := range words{
        l := strings.ToLower(word)
		if l == "fornax"|| l == "sharbert" || l =="kerfuffle"{
			cleanedWords = append(cleanedWords, "****")
			continue
		}
		cleanedWords = append(cleanedWords, word)
	}
	
    res := strings.Join(cleanedWords," ")
	return res
}