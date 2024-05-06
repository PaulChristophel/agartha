package dto

// PageResponse structures the paginated response generically
type PageResponse struct {
	PerPage  int64  `json:"per_page"`  // Total number of items per page that match the query
	NumPages int64  `json:"num_pages"` // Total number of pages that match the query
	Count    int64  `json:"count"`     // Total number of items that match the query
	Next     string `json:"next"`      // URL to the next page of results
	Previous string `json:"previous"`  // URL to the previous page of results
}
