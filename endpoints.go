package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/arnicfil/go_learn_http_chirpy/internal/auth"
	"github.com/arnicfil/go_learn_http_chirpy/internal/database"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type chirpError struct {
	Error string `json:"error"`
}

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Token     string    `json:"token"`
}

func userToResponse(u database.User, token string) UserResponse {
	return UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		Token:     token,
	}
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

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
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

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
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
	type create_user struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	var cu create_user
	if err := decoder.Decode(&cu); err != nil {
		log.Printf("Error decoding request body: %s", err)
		respondWithError(w, http.StatusBadRequest, "Something went wrong")
		return
	}

	hashed_password, err := auth.HashPassword(cu.Password)
	if err != nil {
		log.Printf("Error hashing password: %s", err)
		respondWithError(w, http.StatusBadRequest, "Something went wrong")
		return
	}

	create_user_params := database.CreateUserParams{
		Email:          cu.Email,
		HashedPassword: hashed_password,
	}

	user, err := cfg.queries.CreateUser(r.Context(), create_user_params)
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

	user_secret, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Error while get bearer token: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Something went wrong")
		return
	}

	user_id, err := auth.ValidateJWT(user_secret, cfg.secret)
	if err != nil || user_id != postVal.UserID {
		log.Printf("Error while validating jwt: %v", err)
		respondWithError(w, http.StatusUnauthorized, "Something went wrong")
		return
	}

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

func (cfg *apiConfig) get_chirpsEndpoint(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	chirps, err := cfg.queries.GetAllChirps(r.Context())
	if err != nil {
		log.Printf("Error getting chirps: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	respondWithJSON(w, http.StatusOK, chirps)
}

func (cfg *apiConfig) get_chirpEndpoint(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	id := r.PathValue("chirpID")

	id_uuid, err := uuid.Parse(id)
	if err != nil {
		log.Printf("Error transforming uuid: %s", err)
		respondWithError(w, http.StatusNotFound, "Invalid uuid")
		return
	}

	chirp, err := cfg.queries.GetChirp(r.Context(), id_uuid)
	if err != nil {
		log.Printf("Error getting chirp from uuid: %s", err)
		respondWithError(w, http.StatusNotFound, "Invalid uuid")
		return
	}

	respondWithJSON(w, http.StatusOK, chirp)
}

func (cfg *apiConfig) loginEndpoint(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	type login struct {
		Password           string `json:"password"`
		Email              string `json:"email"`
		Expires_in_seconds int    `json:"expires_in_seconds"`
	}

	decoder := json.NewDecoder(r.Body)
	var l login
	if err := decoder.Decode(&l); err != nil {
		log.Printf("Error decoding request body: %s", err)
		respondWithError(w, http.StatusBadRequest, "Something went wrong")
		return
	}

	if l.Expires_in_seconds == 0 {
		l.Expires_in_seconds = 3600
	}

	user, err := cfg.queries.GetUserWithEmail(r.Context(), l.Email)
	if err != nil {
		log.Printf("Error getting user: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Incorrent password or email")
		return
	}

	match, err := auth.CheckPasswordHash(l.Password, user.HashedPassword)
	if err != nil {
		log.Printf("Error checking password: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Incorrent password or email")
		return
	}

	jwt, err := auth.MakeJWT(user.ID, cfg.secret, time.Second*time.Duration(l.Expires_in_seconds))
	if err != nil {
		log.Printf("Error making jwt: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Incorrent password or email")
		return
	}

	if match {
		respondWithJSON(w, http.StatusOK, userToResponse(user, jwt))
	} else {
		respondWithError(w, http.StatusUnauthorized, "Incorrent password or email")
	}
}
