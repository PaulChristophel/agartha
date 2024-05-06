package dto

import model "github.com/PaulChristophel/agartha/server/model/salt"

// JIDPageResponse structures the paginated response for salt jid queries.
type JIDPageResponse struct {
	Paging  PageResponse `json:"paging"`
	Results []model.JID  `json:"results"` // Array of salt jid results
}
