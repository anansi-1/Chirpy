package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)


type apiConfig struct {
	fileserverHits atomic.Int32
}

// func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler{
// 	cfg.fileserverHits.Add(1)
// 	return  next
// }

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

	apiConfig := apiConfig{
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

	mux.HandleFunc("GET /api/metrics",func(w http.ResponseWriter, r *http.Request) {
		x := apiConfig.getFileServerHits()
		text := fmt.Sprintf("Hits: %d",x)
		fmt.Fprintf(w,text)
	})

	
	mux.HandleFunc("POST /api/reset",func(w http.ResponseWriter, r *http.Request) {
		apiConfig.resetFileServerHits()
	})

	

	srv := &http.Server{  // a struct that describes the server configuration
		Addr: ":" + port,
		Handler: mux,
	}

	log.Printf("Serving on port: %s\n",port)
	log.Fatal(srv.ListenAndServe())

}