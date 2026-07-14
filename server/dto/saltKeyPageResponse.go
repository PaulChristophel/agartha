package dto

import model "github.com/PaulChristophel/agartha/server/model/salt"

type SaltKeyPageResponse struct {
	Paging  PageResponse    `json:"paging"`
	Results []model.SaltKey `json:"results"`
}
