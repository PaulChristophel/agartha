package dto

import model "github.com/PaulChristophel/agartha/server/model/salt/view"

type SaltMinionPageResponse struct {
	Paging  PageResponse       `json:"paging"`
	Results []model.SaltMinion `json:"results"` // Array of salt cache results
}
