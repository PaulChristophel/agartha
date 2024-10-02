package dto

type SaltMinionGrainsKeyResponse struct {
	Paging  PageResponse `json:"paging"`
	Results []string     `json:"results"`
}
