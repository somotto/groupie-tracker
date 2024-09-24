package handlers

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"groupie-trackers/internal/cache"
	"groupie-trackers/internal/models"
)

const (
	artistsPerPage = 10
	cacheDuration  = 60 * time.Minute
)

var (
	artistsCache   = cache.NewCache()
	artistCache    = cache.NewCache()
	locationsCache = cache.NewCache()
	concertsCache  = cache.NewCache()
	templateCache  = make(map[string]*template.Template)
	templateMutex  = &sync.RWMutex{}
)

func RenderErrorPage(w http.ResponseWriter, message string, statusCode int) {
	template, err := getTemplate("error.html")
	if err != nil {
		http.Error(w, "500: Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(statusCode)
	value := models.PageData{ErrorMessage: message}
	template.Execute(w, value)
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	handleArtists(w, r, "")
}

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	handleArtists(w, r, query)
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

func ServeStatic(w http.ResponseWriter, r *http.Request) {
	file := "." + r.URL.Path

	info, err := os.Stat(file)
	if err != nil {
		log.Println("Error getting file info:", err)
		RenderErrorPage(w, "404: File not found", http.StatusNotFound)
		return
	}

	if info.IsDir() {
		log.Println("Error getting file info:", err)
		RenderErrorPage(w, "404: File not Found", http.StatusNotFound)
		return
	}
	http.ServeFile(w, r, file)
}
