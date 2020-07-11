package model

// SearchAPIResults is an object returned by the Mouser API after a search query
type SearchAPIResults struct {
	Errors        []APIError `json:"Errors"`
	SearchResults struct {
		NumberOfResult uint   `json:"NumberOfResult,omitempty"`
		Parts          []Part `json:"Parts"`
	} `json:"SearchResults"`
}
