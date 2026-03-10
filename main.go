package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	initDB()

	go startCleanupWorker()

	mux := http.NewServeMux()

	// static pages
	mux.HandleFunc("/", serveIndex)
	mux.HandleFunc("/p/", servePastePage)

	// api
	mux.HandleFunc("/api/paste", createPasteHandler)
	mux.HandleFunc("/api/paste/", getPasteHandler)
	mux.HandleFunc("/raw/", getRawHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, loggingMiddleware(mux)); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}