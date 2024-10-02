package dto

import model "github.com/PaulChristophel/agartha/server/model/salt/view"

// HighStatePageResponse structures the paginated response for salt highstate queries.
type HighStatePageResponse struct {
	Paging  PageResponse      `json:"paging"`
	Results []model.HighState `json:"results"` // Array of salt highstate results
}
