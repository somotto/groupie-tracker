package handlers

import (
	"encoding/json"
	"fmt"
	"groupie-trackers/internal/models"
	"net/http"
	"sync"
)

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
