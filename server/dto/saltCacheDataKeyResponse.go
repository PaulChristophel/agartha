package dto

type SaltCacheDataKeyResponse struct {
	Paging  PageResponse `json:"paging"`
	Results []string     `json:"results"`
}
