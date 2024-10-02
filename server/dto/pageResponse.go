package dto

// PageResponse structures the paginated response generically
type PageResponse struct {
	PerPage  int64  `json:"per_page" example:"50"`                                                      // Total number of items per page that match the query
	NumPages int64  `json:"num_pages" example:"1626"`                                                   // Total number of pages that match the query
	Count    int64  `json:"count" example:"81286"`                                                      // Total number of items that match the query
	Next     string `json:"next" example:"http://agartha.example.com/api/v1/ex?page=3&per_page=50"`     // URL to the next page of results
	Previous string `json:"previous" example:"http://agartha.example.com/api/v1/ex?page=1&per_page=50"` // URL to the previous page of results
}
