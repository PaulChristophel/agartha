package dto

import model "github.com/PaulChristophel/sylvite/server/model/salt"

// SaltReturnPageResponse structures the paginated response for salt return queries.
type SaltReturnPageResponse struct {
	Paging  PageResponse       `json:"paging"`
	Results []model.SaltReturn `json:"results"` // Array of salt return results
}
