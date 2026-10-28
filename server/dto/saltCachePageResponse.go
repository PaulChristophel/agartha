package dto

import model "github.com/PaulChristophel/sylvite/server/model/salt"

type SaltCachePageResponse struct {
	Paging  PageResponse      `json:"paging"`
	Results []model.SaltCache `json:"results"` // Array of salt cache results
}
