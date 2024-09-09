package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"groupie-trackers/internal/cache"
	"groupie-trackers/internal/models"
)

const (
	artistsPerPage = 10
	cacheDuration  = 5 * time.Minute
)

var (
	artistsCache   = cache.NewCache()
	artistCache    = cache.NewCache()
	locationsCache = cache.NewCache()
	concertsCache  = cache.NewCache()
	templateCache  = make(map[string]*template.Template)
	templateMutex  = &sync.RWMutex{}
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	handleArtists(w, r, "")
}

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	handleArtists(w, r, query)
}

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

	tmpl, err := getTemplate("artist.html")
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

	tmpl, err := getTemplate("index.html")
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, pageData)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func getTemplate(filename string) (*template.Template, error) {
	templateMutex.RLock()
	tmpl, found := templateCache[filename]
	templateMutex.RUnlock()

	if found {
		return tmpl, nil
	}

	templateMutex.Lock()
	defer templateMutex.Unlock()

	tmpl, err := parseTemplate(filename)
	if err != nil {
		return nil, err
	}

	templateCache[filename] = tmpl
	return tmpl, nil
}

func parseTemplate(filename string) (*template.Template, error) {
	tmplPath := filepath.Join("internal", "templates", filename)
	return template.New(filename).Funcs(template.FuncMap{
		"subtract": func(a, b int) int { return a - b },
		"add":      func(a, b int) int { return a + b },
		"sequence": func(n int) []int {
			seq := make([]int, n)
			for i := range seq {
				seq[i] = i + 1
			}
			return seq
		},
	}).ParseFiles(tmplPath)
}

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

func fetchArtistDetails(id int) (models.Artist, error) {
	cacheKey := fmt.Sprintf("artist_%d", id)
	if cachedArtist, found := artistCache.Get(cacheKey); found {
		return cachedArtist.(models.Artist), nil
	}

	var artist models.Artist
	var relations struct {
		DatesLocations map[string][]string `json:"datesLocations"`
	}

	artistChan := make(chan error, 1)
	relationsChan := make(chan error, 1)

	go func() {
		resp, err := http.Get(fmt.Sprintf("https://groupietrackers.herokuapp.com/api/artists/%d", id))
		if err != nil {
			artistChan <- err
			return
		}
		defer resp.Body.Close()
		err = json.NewDecoder(resp.Body).Decode(&artist)
		artistChan <- err
	}()

	go func() {
		resp, err := http.Get(fmt.Sprintf("https://groupietrackers.herokuapp.com/api/relation/%d", id))
		if err != nil {
			relationsChan <- err
			return
		}
		defer resp.Body.Close()
		err = json.NewDecoder(resp.Body).Decode(&relations)
		relationsChan <- err
	}()

	if err := <-artistChan; err != nil {
		return models.Artist{}, err
	}
	if err := <-relationsChan; err != nil {
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

func ConcertsHandler(w http.ResponseWriter, r *http.Request) {
	artistConcerts, err := fetchArtistConcerts()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching artist concerts: %v", err), http.StatusInternalServerError)
		return
	}

	tmpl, err := getTemplate("concerts.html")
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := struct {
		ArtistConcerts map[string]map[string][]string
	}{
		ArtistConcerts: artistConcerts,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func fetchArtistConcerts() (map[string]map[string][]string, error) {
	if cachedConcerts, found := concertsCache.Get("artistConcerts"); found {
		return cachedConcerts.(map[string]map[string][]string), nil
	}

	artists, err := fetchArtists()
	if err != nil {
		return nil, fmt.Errorf("error fetching artists: %v", err)
	}

	artistConcerts := make(map[string]map[string][]string)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, artist := range artists {
		wg.Add(1)
		go func(a models.Artist) {
			defer wg.Done()
			artistDetails, err := fetchArtistDetails(a.ID)
			if err != nil {
				fmt.Printf("Error fetching details for artist %s: %v\n", a.Name, err)
				return
			}
			mu.Lock()
			artistConcerts[a.Name] = artistDetails.RelationsData
			mu.Unlock()
		}(artist)
	}

	wg.Wait()

	concertsCache.Set("artistConcerts", artistConcerts, cacheDuration)
	return artistConcerts, nil
}

func LocationsHandler(w http.ResponseWriter, r *http.Request) {
	artistLocations, err := fetchArtistLocations()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching artist locations: %v", err), http.StatusInternalServerError)
		return
	}

	tmpl, err := getTemplate("locations.html")
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

func fetchArtistLocations() (map[string][]string, error) {
	if cachedLocations, found := locationsCache.Get("artistLocations"); found {
		return cachedLocations.(map[string][]string), nil
	}

	artistConcerts, err := fetchArtistConcerts()
	if err != nil {
		return nil, fmt.Errorf("error fetching artist concerts: %v", err)
	}

	artistLocations := make(map[string][]string)
	for artistName, concerts := range artistConcerts {
		var locations []string
		for location := range concerts {
			locations = append(locations, location)
		}
		artistLocations[artistName] = locations
	}

	locationsCache.Set("artistLocations", artistLocations, cacheDuration)
	return artistLocations, nil
}

func DatesHandler(w http.ResponseWriter, r *http.Request) {
	artistConcerts, err := fetchArtistConcerts()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching artist concerts: %v", err), http.StatusInternalServerError)
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
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := struct {
		ArtistDates map[string][]string
	}{
		ArtistDates: artistDates,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
