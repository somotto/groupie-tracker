package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"groupie-trackers/internal/cache"
	"groupie-trackers/internal/models"
)

// define the number of artists to be shown per page and the duration of one cache.
const (
	artistsPerPage = 10
	cacheDuration  = 5 * time.Minute
)

// Creates caches for storing lists of artists and individual artist details to reduce repeated API calls.
var (
	artistsCache = cache.NewCache()
	artistCache  = cache.NewCache()
	locationsCache = cache.NewCache()
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	handleArtists(w, r, "")
}

// Handles search queries for artists based on the `q` parameter in the URL.
func SearchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	handleArtists(w, r, query)
}

// It fetches artist data, renders the artist.html template with the artist's details, and serves it to the client.
func ArtistHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid artist ID", http.StatusBadRequest)
		return
	}

	artist, err := fetchArtistDetails(id)
	if err != nil {
		http.Error(w, "Error fetching artist details", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles(filepath.Join("internal", "templates", "artist.html"))
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := models.ArtistPageData{
		Artist: artist,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func handleArtists(w http.ResponseWriter, r *http.Request, searchQuery string) {
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 1
	}

	sortBy := r.URL.Query().Get("sort")

	artists, err := fetchArtists()
	if err != nil {
		http.Error(w, "Error fetching artists", http.StatusInternalServerError)
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

	tmpl, err := parseTemplate("index.html")
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, pageData)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// Reads and parses HTML templates from files, allowing custom functions like subtract, add, and sequence to be used within temp
func parseTemplate(filename string) (*template.Template, error) {
	tmplPath := filepath.Join("internal", "templates", filename)

	// Check if the file exists
	if _, err := os.Stat(tmplPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("template file does not exist: %s", tmplPath)
	}

	// Try to read the file content
	content, err := os.ReadFile(tmplPath)
	if err != nil {
		return nil, fmt.Errorf("error reading template file: %v", err)
	}

	// Parse the template
	tmpl, err := template.New(filename).Funcs(template.FuncMap{
		"subtract": func(a, b int) int { return a - b },
		"add":      func(a, b int) int { return a + b },
		"sequence": func(n int) []int {
			seq := make([]int, n)
			for i := range seq {
				seq[i] = i + 1
			}
			return seq
		},
	}).Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("error parsing template content: %v", err)
	}

	return tmpl, nil
}

// Fetches the list of artists from an external API, caches the result, and returns it. If the data is already cached, it uses the cached version.
func fetchArtists() ([]models.Artist, error) {
	if cachedArtists, found := artistsCache.Get("artists"); found {
		return cachedArtists.([]models.Artist), nil
	}

	resp, err := http.Get("https://groupietrackers.herokuapp.com/api/artists")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var artists []models.Artist
	err = json.NewDecoder(resp.Body).Decode(&artists)
	if err != nil {
		return nil, err
	}

	artistsCache.Set("artists", artists, cacheDuration)
	return artists, nil
}

// Fetches details for a specific artist, including related data, and caches the result. Uses cached data if available.
func fetchArtistDetails(id int) (models.Artist, error) {
	cacheKey := fmt.Sprintf("artist_%d", id)
	if cachedArtist, found := artistCache.Get(cacheKey); found {
		return cachedArtist.(models.Artist), nil
	}

	resp, err := http.Get(fmt.Sprintf("https://groupietrackers.herokuapp.com/api/artists/%d", id))
	if err != nil {
		return models.Artist{}, err
	}
	defer resp.Body.Close()

	var artist models.Artist
	err = json.NewDecoder(resp.Body).Decode(&artist)
	if err != nil {
		return models.Artist{}, err
	}

	relationsResp, err := http.Get(artist.Relations)
	if err != nil {
		return models.Artist{}, err
	}
	defer relationsResp.Body.Close()

	var relations struct {
		DatesLocations map[string][]string `json:"datesLocations"`
	}
	err = json.NewDecoder(relationsResp.Body).Decode(&relations)
	if err != nil {
		return models.Artist{}, err
	}

	artist.RelationsData = relations.DatesLocations

	artistCache.Set(cacheKey, artist, cacheDuration)
	return artist, nil
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
			return artists[i].FirstAlbum < artists[j].FirstAlbum
		})
	}
}

// It fetches all artist data, calculates statistics, and renders them in the statistics.html template.
func StatisticsHandler(w http.ResponseWriter, r *http.Request) {
	artists, err := fetchArtists()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	stats := calculateStatistics(artists)

	tmpl, err := template.ParseFiles(filepath.Join("internal", "templates", "statistics.html"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, stats)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func calculateStatistics(artists []models.Artist) models.Statistics {
	var stats models.Statistics
	creationYears := make(map[int]int)
	memberCounts := make(map[int]int)

	for _, artist := range artists {
		creationYears[artist.CreationDate]++
		memberCounts[len(artist.Members)]++
	}

	stats.CreationYearData = creationYears
	stats.MemberCountData = memberCounts

	return stats
}

func LocationsHandler(w http.ResponseWriter, r *http.Request) {
    artistLocations, err := fetchArtistLocations()
    if err != nil {
        http.Error(w, fmt.Sprintf("Error fetching artist locations: %v", err), http.StatusInternalServerError)
        return
    }

    tmpl, err := template.ParseFiles(filepath.Join("internal", "templates", "locations.html"))
    if err != nil {
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    data := struct {
        ArtistLocations map[string][]string
    }{
        ArtistLocations: artistLocations,
    }

    err = tmpl.Execute(w, data)
    if err != nil {
        http.Error(w, "Internal server error", http.StatusInternalServerError)
    }
}


// Fetches locations from the API and caches them
func fetchArtistLocations() (map[string][]string, error) {
    if cachedLocations, found := locationsCache.Get("artistLocations"); found {
        return cachedLocations.(map[string][]string), nil
    }

    artists, err := fetchArtists()
    if err != nil {
        return nil, fmt.Errorf("error fetching artists: %v", err)
    }

    artistLocations := make(map[string][]string)

    for _, artist := range artists {
        relations, err := fetchRelations(artist.Relations)
        if err != nil {
            return nil, fmt.Errorf("error fetching relations for artist %s: %v", artist.Name, err)
        }

        var locations []string
        for location := range relations.DatesLocations {
            locations = append(locations, location)
        }

        artistLocations[artist.Name] = locations
    }

    locationsCache.Set("artistLocations", artistLocations, cacheDuration)
    return artistLocations, nil
}

func fetchRelations(relationsURL string) (*models.Relations, error) {
    resp, err := http.Get(relationsURL)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var relations models.Relations
    err = json.NewDecoder(resp.Body).Decode(&relations)
    if err != nil {
        return nil, err
    }

    return &relations, nil
}