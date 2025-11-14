package main

import _ "github.com/lib/pq"

import (
	"database/sql"
	"fmt"
	"github.com/arnicfil/go_learn_http_chirpy/internal/database"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	queries        *database.Queries
	platform       string
	secret         string
}

func run() error {
	err := godotenv.Load()
	if err != nil {
		return err
	}

	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return err
	}
	dbQueries := database.New(db)

	apiCfg := apiConfig{
		queries:  dbQueries,
		platform: os.Getenv("PLATFORM"),
		secret:   os.Getenv("SECRET"),
	}

	filepathRoot, err := os.Getwd()
	if err != nil {
		return err
	}

	fileSystem := http.FileServer(http.Dir((filepathRoot)))

	DefaultServeMux := http.NewServeMux()
	DefaultServeMux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", fileSystem)))
	DefaultServeMux.HandleFunc("GET /api/healthz", readinessEndpoint)
	DefaultServeMux.HandleFunc("GET /admin/metrics", apiCfg.hitsEndpoint)
	DefaultServeMux.HandleFunc("POST /admin/reset", apiCfg.resetEndpoint)
	DefaultServeMux.HandleFunc("POST /api/users", apiCfg.create_userEndpoint)
	DefaultServeMux.HandleFunc("POST /api/chirps", apiCfg.create_chirpEndpoint)
	DefaultServeMux.HandleFunc("GET /api/chirps", apiCfg.get_chirpsEndpoint)
	DefaultServeMux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.get_chirpEndpoint)
	DefaultServeMux.HandleFunc("POST /api/login", apiCfg.loginEndpoint)

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
