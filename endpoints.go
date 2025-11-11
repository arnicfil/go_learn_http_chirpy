package main

import _ "github.com/lib/pq"

import (
	"encoding/json"
	"fmt"
	"github.com/arnicfil/go_learn_http_chirpy/internal/database"
	"io"
	"log"
	"net/http"
	"slices"
	"strings"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	queries        *database.Queries
}

type chirpValue struct {
	Body string `json:"body"`
}

type chirpError struct {
	Error string `json:"error"`
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
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

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	hitsValue := cfg.fileserverHits.Load()
	w.Write(fmt.Appendf(nil, "<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", hitsValue))
}

func (cfg *apiConfig) resetEndpoint(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)

	defer r.Body.Close()
	w.WriteHeader(http.StatusOK)
}

func respondWithError(w http.ResponseWriter, status int, msg string) {
	body, err := json.Marshal(chirpError{Error: msg})
	if err != nil {
		log.Printf("Json marshal failed: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "aplication/json; charset=utf-8")
	w.WriteHeader(status)
	w.Write(body)
}

func respondWithJSON(w http.ResponseWriter, status int, vals any) {
	body, err := json.Marshal(vals)
	if err != nil {
		log.Printf("Json marshal failed: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "aplication/json; charset=utf-8")
	w.WriteHeader(status)
	w.Write(body)
}

func validate_chirpEndpoint(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	var chirpval chirpValue
	if err := decoder.Decode(&chirpval); err != nil {
		log.Printf("Error decoding request body: %s", err)
		respondWithError(w, http.StatusBadRequest, "Something went wrong")
		return
	}

	if len(chirpval.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	chirpval.Body = cleanString(chirpval.Body)
	respondWithJSON(w, http.StatusOK, chirpval)
}

func cleanString(s string) (cleanS string) {
	var badWords = []string{"kerfuffle", "sharbert", "fornax"}
	words := strings.Fields(s)
	for i, word := range words {
		if slices.Contains(badWords, word) {
			words[i] = "****"
		}
	}

	return strings.Join(words, " ")
}
