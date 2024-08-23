package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"groupie-trackers/internal/handlers"
)

func main() {
	// Change the working directory to the project root
	err := os.Chdir(filepath.Join("..", ".."))
	if err != nil {
		log.Fatal(err)
	}

	// Routes incoming requests to different handlers based on the request URL.
	mux := http.NewServeMux()

	// Serve static files
	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Register handlers
	mux.HandleFunc("/", handlers.HomeHandler)
	mux.HandleFunc("/search", handlers.SearchHandler)
	mux.HandleFunc("/artist", handlers.ArtistHandler)
	mux.HandleFunc("/statistics", handlers.StatisticsHandler)

	log.Println("Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
