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
	http.HandleFunc("/", handler)
	http.HandleFunc("/static/", handlers.ServeStatic)

	port := ":8080"

	fmt.Printf("Server is running on http://localhost%s", port)

	log.Fatal(http.ListenAndServe(port, nil))
}

// Handle different URL paths
func handler(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		handlers.HomeHandler(w, r)
	case "/search":
		handlers.SearchHandler(w, r)
	case "/artist":
		handlers.ArtistHandler(w, r)
	case "/dates":
		handlers.DatesHandler(w, r)
	case "/concerts":
		handlers.ConcertsHandler(w, r)
	case "/locations":
		handlers.LocationsHandler(w, r)
	default:
		handlers.RenderErrorPage(w, "404: Page not found", http.StatusNotFound)
		return
	}
}
