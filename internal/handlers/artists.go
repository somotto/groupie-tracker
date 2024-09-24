package handlers

import (
	"groupie-trackers/internal/models"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

func ArtistHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)

	if err != nil || !(id >= 1 && id <= 52) {
		log.Println("Invalid artist id: ", id)
		RenderErrorPage(w, "400: Bad Request", http.StatusBadRequest)
		return
	}

	artist, err := fetchArtistDetails(id)
	if err != nil {
		log.Println("Error fetching artist details: ", err)
		RenderErrorPage(w, "500: Internal server error", http.StatusInternalServerError)
		return
	}

	tmpl, err := getTemplate("artist.html")
	if err != nil {
		log.Println("Error getting template: ", err)
		RenderErrorPage(w, "500: Internal server error", http.StatusInternalServerError)
		return
	}

	data := models.ArtistPageData{
		Artist: artist,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Println("Error executing template: ", err)
		RenderErrorPage(w, "500: Internal server error", http.StatusInternalServerError)
	}
}

func handleArtists(w http.ResponseWriter, r *http.Request, searchQuery string) {
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 || page > 6 {
		page = 1
	}

	sortBy := r.URL.Query().Get("sort")

	artists, err := fetchArtists()
	if err != nil {
		log.Println("Error fetching artists: ", err)
		http.Error(w, "500: Internal server error", http.StatusInternalServerError)
		return
	}

	if searchQuery != "" {
		artists = filterArtists(artists, searchQuery)
	}

	sortArtists(artists, sortBy)

	totalPages := (len(artists) + artistsPerPage - 1) / artistsPerPage
	startIndex := (page - 1) * artistsPerPage
	endIndex := startIndex + artistsPerPage
	if endIndex > len(artists) {
		endIndex = len(artists)
	}

	pageData := models.PageData{
		Artists:     artists[startIndex:endIndex],
		TotalPages:  totalPages,
		CurrentPage: page,
		SearchQuery: searchQuery,
		SortBy:      sortBy,
	}

	tmpl, err := getTemplate("index.html")
	if err != nil {
		log.Println("Error getting template:", err)
		RenderErrorPage(w, "500: Internal server error", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, pageData)
	if err != nil {
		log.Println("Error executing template:", err)
		RenderErrorPage(w, "500: Internal server error", http.StatusInternalServerError)
	}
}

func sortArtists(artists []models.Artist, sortBy string) {
	switch sortBy {
	case "name":
		sort.Slice(artists, func(i, j int) bool {
			return artists[i].Name < artists[j].Name
		})
	case "creationDate":
		sort.Slice(artists, func(i, j int) bool {
			return artists[i].CreationDate < artists[j].CreationDate
		})
	case "firstAlbum":
		sort.Slice(artists, func(i, j int) bool {
			dateI, errI := time.Parse("02-01-2006", artists[i].FirstAlbum)
			dateJ, errJ := time.Parse("02-01-2006", artists[j].FirstAlbum)
			if errI != nil || errJ != nil {
				log.Println("Error parsing date:", errI, errJ)
				return artists[i].FirstAlbum < artists[j].FirstAlbum
			}
			return dateI.Before(dateJ)
		})
	}
}

func filterArtists(artists []models.Artist, query string) []models.Artist {
	query = strings.ToLower(query)
	var filtered []models.Artist
	for _, artist := range artists {
		if strings.Contains(strings.ToLower(artist.Name), query) {
			filtered = append(filtered, artist)
		}
	}
	return filtered
}
