package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetrixsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func readinessEndpoint(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(bodyBytes)
}

func (cfg *apiConfig) hitsEndpoint(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	hitsValue := cfg.fileserverHits.Load()
	w.Write(fmt.Appendf(nil, "Hits: %d", hitsValue))
}

func (cfg *apiConfig) resetEndpoint(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)

	defer r.Body.Close()
	w.WriteHeader(http.StatusOK)
}

func run() error {
	var apiCfg apiConfig
	filepathRoot, err := os.Getwd()
	if err != nil {
		return err
	}

	fileSystem := http.FileServer(http.Dir((filepathRoot)))

	DefaultServeMux := http.NewServeMux()
	DefaultServeMux.Handle("/app/", apiCfg.middlewareMetrixsInc(http.StripPrefix("/app/", fileSystem)))
	DefaultServeMux.HandleFunc("/healthz/", readinessEndpoint)
	DefaultServeMux.HandleFunc("/metrics/", apiCfg.hitsEndpoint)
	DefaultServeMux.HandleFunc("/reset/", apiCfg.resetEndpoint)

	port := "8080"
	s := &http.Server{
		Addr:    ":" + port,
		Handler: DefaultServeMux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(s.ListenAndServe())
	return nil
}

func main() {
	err := run()
	if err != nil {
		fmt.Printf("Error while running: %v", err)
		os.Exit(1)
	}
}
