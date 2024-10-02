package dto

type SaltMinionPillarKeyResponse struct {
	Paging  PageResponse `json:"paging"`
	Results []string     `json:"results"`
}
