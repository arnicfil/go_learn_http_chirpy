package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

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

func run() error {
	filepathRoot, err := os.Getwd()
	if err != nil {
		return err
	}

	fileSystem := http.FileServer(http.Dir((filepathRoot)))

	DefaultServeMux := http.NewServeMux()
	DefaultServeMux.Handle("/app/", http.StripPrefix("/app/", fileSystem))
	DefaultServeMux.HandleFunc("/healthz/", readinessEndpoint)

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
