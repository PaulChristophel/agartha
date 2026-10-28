package dto

import model "github.com/PaulChristophel/sylvite/server/model/salt/materializedView"

// conformityPageResponse structures the paginated response for salt conformity queries.
type ConformityPageResponse struct {
	Paging  PageResponse       `json:"paging"`
	Results []model.Conformity `json:"results"` // Array of salt conformity results
}
