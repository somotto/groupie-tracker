package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"groupie-trackers/internal/handlers"
)

func main() {
	if len(os.Args) != 1 {
		return
	}
	// Change the working directory to the project root
	err := os.Chdir(filepath.Join("..", ".."))
	if err != nil {
		log.Fatal(err)
	}

	// Register handlers
	http.HandleFunc("/static/", handlers.ServeStatic)
	http.HandleFunc("/", handlers.HomeHandler)
	http.HandleFunc("/search", handlers.SearchHandler)
	http.HandleFunc("/artist", handlers.ArtistHandler)
	http.HandleFunc("/dates", handlers.DatesHandler)
	http.HandleFunc("/concerts", handlers.ConcertsHandler)
	http.HandleFunc("/locations", handlers.LocationsHandler)

	port := ":8000"

	fmt.Printf("Server is running on http://localhost%s", port)

	log.Fatal(http.ListenAndServe(port, nil))
}
