package dto

import model "github.com/PaulChristophel/agartha/server/model/salt"

// SaltEventPageResponse structures the paginated response for salt event queries.
type SaltEventPageResponse struct {
	Paging  PageResponse      `json:"paging"`
	Results []model.SaltEvent `json:"results"` // Array of salt event results
}
