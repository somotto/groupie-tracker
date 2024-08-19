package models

type Artist struct {
	ID             int                 `json:"id"`
	Image          string              `json:"image"`
	Name           string              `json:"name"`
	Members        []string            `json:"members"`
	CreationDate   int                 `json:"creationDate"`
	FirstAlbum     string              `json:"firstAlbum"`
	Locations      string              `json:"locations"`
	ConcertDates   string              `json:"concertDates"`
	Relations      string              `json:"relations"`
	RelationsData  map[string][]string `json:"-"`
}

type PageData struct {
	Artists     []Artist
	TotalPages  int
	CurrentPage int
	SearchQuery string
	SortBy      string
}

type ArtistPageData struct {
	Artist Artist
}

type Statistics struct {
	CreationYearData map[int]int
	MemberCountData  map[int]int
}