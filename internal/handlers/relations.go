package handlers

import (
	"log"
	"net/http"
)

func DatesHandler(w http.ResponseWriter, r *http.Request) {
	artistConcerts, err := fetchArtistConcerts()
	if err != nil {
		log.Println("Error fetching artist concerts", err)
		RenderErrorPage(w, "500: Internal server error", http.StatusInternalServerError)
		return
	}

	artistDates := make(map[string][]string)
	for artistName, concertData := range artistConcerts {
		var allDates []string
		for _, dates := range concertData {
			allDates = append(allDates, dates...)
		}
		artistDates[artistName] = allDates
	}

	tmpl, err := getTemplate("dates.html")
	if err != nil {
		log.Println("Error getting template:", err)
		RenderErrorPage(w, "500: Internal server error", http.StatusInternalServerError)
		return
	}

	data := struct {
		ArtistDates map[string][]string
	}{
		ArtistDates: artistDates,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Println("Error executing template:", err)
		RenderErrorPage(w, "500: Internal server error", http.StatusInternalServerError)
	}
}

func LocationsHandler(w http.ResponseWriter, r *http.Request) {
	artistLocations, err := fetchArtistLocations()
	if err != nil {
		log.Println("Error fetching artist locations", err)
		RenderErrorPage(w, "500: Internal server error", http.StatusInternalServerError)
		return
	}

	tmpl, err := getTemplate("locations.html")
	if err != nil {
		log.Println("Error getting template:", err)
		RenderErrorPage(w, "500: Internal server error", http.StatusInternalServerError)
		return
	}

	data := struct {
		ArtistLocations map[string][]string
	}{
		ArtistLocations: artistLocations,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Println("Error executing template:", err)
		RenderErrorPage(w, "500: Internal server error", http.StatusInternalServerError)
		return
	}
}

func ConcertsHandler(w http.ResponseWriter, r *http.Request) {
	artistConcerts, err := fetchArtistConcerts()
	if err != nil {
		log.Println("Error fetching artist", err)
		RenderErrorPage(w, "500: Internal server error", http.StatusInternalServerError)
		return
	}

	tmpl, err := getTemplate("concerts.html")
	if err != nil {
		log.Println("Error getting template:", err)
		RenderErrorPage(w, "500: Internal server error", http.StatusInternalServerError)
		return
	}

	data := struct {
		ArtistConcerts map[string]map[string][]string
	}{
		ArtistConcerts: artistConcerts,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Println("Error executing template:", err)
		RenderErrorPage(w, "500: Internal server error", http.StatusInternalServerError)
		return
	}
}
