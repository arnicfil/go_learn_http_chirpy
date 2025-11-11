package main

import _ "github.com/lib/pq"

import (
	"encoding/json"
	"fmt"
	"github.com/arnicfil/go_learn_http_chirpy/internal/database"
	"github.com/google/uuid"
	"io"
	"log"
	"net/http"
	"slices"
	"strings"
	"time"
)

type chirpValue struct {
	Body string `json:"body"`
}

type chirpError struct {
	Error string `json:"error"`
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
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
	defer r.Body.Close()
	if cfg.platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	cfg.fileserverHits.Store(0)
	err := cfg.queries.DeleteUsers(r.Context())
	if err != nil {
		log.Printf("Error deleting user: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

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

func (cfg *apiConfig) create_userEndpoint(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	type email struct {
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	var mail email
	if err := decoder.Decode(&mail); err != nil {
		log.Printf("Error decoding request body: %s", err)
		respondWithError(w, http.StatusBadRequest, "Something went wrong")
		return
	}

	user, err := cfg.queries.CreateUser(r.Context(), mail.Email)
	if err != nil {
		log.Printf("Error creating user: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	respondWithJSON(w, http.StatusCreated, user)
}

func (cfg *apiConfig) create_chirpEndpoint(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	type payload struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	decoder := json.NewDecoder(r.Body)
	var postVal payload
	if err := decoder.Decode(&postVal); err != nil {
		log.Printf("Error decoding request body: %s", err)
		respondWithError(w, http.StatusBadRequest, "Something went wrong")
		return
	}

	if len(postVal.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	postVal.Body = cleanString(postVal.Body)

	chirp, err := cfg.queries.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   postVal.Body,
		UserID: postVal.UserID,
	})
	if err != nil {
		log.Printf("Error creating chirp: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	respondWithJSON(w, http.StatusCreated, chirp)
}
