package main

import "net/http"


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
