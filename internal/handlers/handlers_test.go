package handlers

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"groupie-trackers/internal/models"
)

func TestMain(m *testing.M) {
	// Set the working directory to the project root
	if err := os.Chdir("../../"); err != nil {
		log.Fatalf("could not change working directory: %v", err)
	}

	// Run the tests
	os.Exit(m.Run())
}

type TestSuite struct {
	t *testing.T
}

func NewTestSuite(t *testing.T) *TestSuite {
	return &TestSuite{t: t}
}

func TestArtistHandlers(t *testing.T) {
	suite := NewTestSuite(t)
	suite.TestSortArtists()
	suite.TestFilterArtists()
	suite.TestAllHandlers()
	suite.TestInvalidHandlers()
	suite.TestTemplateRendering()
}

func (ts *TestSuite) TestAllHandlers() {
	ts.testHandler("/", HomeHandler)
	ts.testHandler("/search?q=1", SearchHandler)
	ts.testHandler("/artist?id=1", ArtistHandler)
	ts.testHandler("/dates", DatesHandler)
	ts.testHandler("/concerts", ConcertsHandler)
	ts.testHandler("/locations", LocationsHandler)
}

func (ts *TestSuite) TestInvalidHandlers() {
	ts.testEdgeCase("/artist", ArtistHandler, http.StatusBadRequest)            // Missing id parameter
	ts.testEdgeCase("/artist?id=invalid", ArtistHandler, http.StatusBadRequest) // Invalid id parameter
}

func (ts *TestSuite) TestTemplateRendering() {
	ts.testTemplateRendering("/artist?id=1", ArtistHandler)
	ts.testTemplateRendering("/concerts", ConcertsHandler)
	ts.testTemplateRendering("/locations", LocationsHandler)
	ts.testTemplateRendering("/dates", DatesHandler)
}

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

func (ts *TestSuite) testTemplateRendering(path string, handlerFunc http.HandlerFunc) {
	// Create a request to simulate an HTTP GET to the root path
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		ts.t.Fatal(err)
	}

	// Create a ResponseRecorder to capture the handler's response
	rr := httptest.NewRecorder()
	handlerFunc(rr, req)

	// Check the content type (expecting text/html)
	contentType := rr.Header().Get("Content-Type")
	want := "text/html; charset=utf-8"
	if contentType != want {
		ts.t.Errorf("handler did not set the Content-Type header correctly: got %v, want %v", contentType, want)
	}
}

func (ts *TestSuite) testHandler(path string, handlerFunc http.HandlerFunc) {
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		ts.t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handlerFunc(rr, req)

	if status := rr.Code; status != http.StatusOK {
		ts.t.Errorf("handler returned wrong status code: got %v, want %v",
			status, http.StatusOK)
	}
}

func (ts *TestSuite) testEdgeCase(path string, handlerFunc http.HandlerFunc, expectedStatus int) {
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		ts.t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handlerFunc(rr, req)

	if status := rr.Code; status != expectedStatus {
		ts.t.Errorf("handler returned wrong status code for edge case: got %v, want %v",
			status, expectedStatus)
	}
}
