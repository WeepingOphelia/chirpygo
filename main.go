package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileServerHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		cfg.fileServerHits.Add(1)
		next.ServeHTTP(w, req)
	})
}

func (a *apiConfig) handlerMetrics(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	someShit := fmt.Sprintf("Hits: %v", a.fileServerHits.Load())
	w.Write([]byte(someShit))
}

func (a *apiConfig) handlerReset(w http.ResponseWriter, req *http.Request) {
	a.fileServerHits.Store(0)
}

func main() {
	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir("."))
	nude := http.StripPrefix("/app", fileServer)
	newAPIConfig := apiConfig{}
	mux.Handle("/app/", newAPIConfig.middlewareMetricsInc(nude))
	ready := func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	}

	mux.HandleFunc("GET /api/healthz", ready)
	mux.HandleFunc("GET /api/metrics", newAPIConfig.handlerMetrics)
	mux.HandleFunc("POST /api/reset", newAPIConfig.handlerReset)
	newServe := http.Server{Handler: mux, Addr: ":8080"}
	newServe.ListenAndServe()
}
