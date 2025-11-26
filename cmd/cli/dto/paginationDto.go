package dto

type PaginatedMangaResponse struct {
	Data       []MangaResponse `json:"data"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	Total      int64           `json:"total"`
	TotalPages int             `json:"total_pages"`
}

func NewPaginatedMangaResponse(data []MangaResponse, page, pageSize int, total int64) PaginatedMangaResponse {
	totalPages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		totalPages++
	}

	return PaginatedMangaResponse{
		Data:       data,
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
	}
}
