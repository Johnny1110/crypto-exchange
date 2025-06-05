package dto

type PaginationResp[T any] struct {
	Total       int64 `json:"total"`
	TotalPages  int64 `json:"total_page"`
	CurrentPage int64 `json:"current_page"`
	PageSize    int64 `json:"page_size"`
	HasNext     bool  `json:"has_next"`
	HasPrev     bool  `json:"has_prev"`
	Result      []T   `json:"result"`
}
