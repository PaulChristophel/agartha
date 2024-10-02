package dto

// SaltReturnPageResponse structures the paginated response for salt return queries.
type SaltReturnFunPageResponse struct {
	Paging  PageResponse `json:"paging"`
	Results []string     `json:"results"` // Array of salt functions
}
