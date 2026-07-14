package dto

import model "github.com/PaulChristophel/agartha/server/model/salt"

// SaltKeyPageResponse is a paginated salt_keys API response.
type SaltKeyPageResponse struct {
	Paging  PageResponse    `json:"paging"`
	Results []model.SaltKey `json:"results"`
}
