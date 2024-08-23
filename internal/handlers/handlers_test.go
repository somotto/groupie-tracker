package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"groupie-trackers/internal/models"
)

type TestSuite struct {
	t *testing.T
}

func NewTestSuite(t *testing.T) *TestSuite {
	return &TestSuite{t: t}
}

func TestArtistHandler(t *testing.T) {
	suite := NewTestSuite(t)
	suite.TestSortArtists()
	suite.TestFilterArtists()
	suite.TestArtistHandler()
	suite.TestSearchHandler()
	suite.TestHomeHandler()
	// suite.TestParseTemplate()
	suite.TestTemplateRendering()
}

// func (ts *TestSuite) TestParseTemplate() {
// 	tmpl, err := template.ParseFiles(filepath.Join("../", "templates", "artist.html"))
// 	// tmpl, err := parseTemplate("../templates/index.html")
// 	if err != nil {
// 		ts.t.Error(err)
// 	}
// 	if tmpl == nil {
// 		ts.t.Error("Expected non-nil template, got nil")
// 	}
// }

func (ts *TestSuite) TestSortArtists() {
	artists := []models.Artist{
		{Name: "C Artist", CreationDate: 2000, FirstAlbum: "2002-01-01"},
		{Name: "A Artist", CreationDate: 2010, FirstAlbum: "2012-01-01"},
		{Name: "B Artist", CreationDate: 2005, FirstAlbum: "2007-01-01"},
	}

	sortArtists(artists, "creationDate")
	if artists[0].CreationDate != 2000 {
		ts.t.Errorf("Expected sorted artists, got %v, want %v", artists[0].CreationDate, 2000)
	}
	sortArtists(artists, "name")
	if artists[0].Name != "A Artist" {
		ts.t.Errorf("Expected sorted artists, got %v, want %v", artists[0].Name, "C Artist")
	}
	sortArtists(artists, "firstAlbum")
	if artists[0].FirstAlbum != "2002-01-01" {
		ts.t.Errorf("Expected sorted artists, got %v, want %v", artists[0].FirstAlbum, "2002-01-01")
	}
}

func (ts *TestSuite) TestFilterArtists() {
	artists := []models.Artist{
		{Name: "Test Artist 1"},
		{Name: "Test Artist 2"},
		{Name: "Test Artist 3"},
	}

	filtered := filterArtists(artists, "test")
	if len(filtered) != 3 {
		ts.t.Error("Expected 3 filtered artists, got", len(filtered))
	}
}

func (ts *TestSuite) TestArtistHandler() {
	req, err := http.NewRequest("GET", "/artist?id=1", nil)
	if err != nil {
		ts.t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	ArtistHandler(rr, req)
	if status := rr.Code; status != http.StatusInternalServerError {
		ts.t.Errorf("handler returned wrong status code: got %v, want %v",
			status, http.StatusInternalServerError)
	}
}

func (ts *TestSuite) TestSearchHandler() {
	req, err := http.NewRequest("GET", "/search?q=test", nil)
	if err != nil {
		ts.t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	SearchHandler(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		ts.t.Errorf("handler returned wrong status code: got %v, want %v",
			status, http.StatusInternalServerError)
	}
}

func (ts *TestSuite) TestHomeHandler() {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		ts.t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	HomeHandler(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		ts.t.Errorf("handler returned wrong status code: got %v, want %v",
			status, http.StatusInternalServerError)
	}
}

func (ts *TestSuite) TestTemplateRendering() {
	// Create a request to simulate an HTTP GET to the root path
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		ts.t.Fatal(err)
	}

	// Create a ResponseRecorder to capture the handler's response
	rr := httptest.NewRecorder()
	ArtistHandler(rr, req)

	// Check the content type (expecting text/html)
	contentType := rr.Header().Get("Content-Type")
	want := "text/plain; charset=utf-8"
	if contentType != want {
		ts.t.Errorf("handler did not set the Content-Type header correctly: got %v, want %v", contentType, want)
	}
}
