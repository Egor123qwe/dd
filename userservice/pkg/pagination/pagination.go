package pagination

import (
	"strconv"
)

const (
	DefaultPageSize = 30
	DefaultPageNum  = 1
)

type Req struct {
	PageSize   int
	PageNumber int
}

func (r *Req) GetOffset() int {
	if r == nil {
		return 0
	}
	return (r.PageNumber - 1) * r.PageSize
}

type Resp struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalItems int `json:"total_items"`
}

func New(pageSizeQuery, pageNumberQuery string) *Req {
	pageSizeStr := pageSizeQuery
	pageNumberStr := pageNumberQuery

	if pageSizeStr == "" && pageNumberStr == "" {
		return nil
	}

	var pageSize, pageNumber int
	var hasPageSize, hasPageNumber bool

	if pageSizeStr != "" {
		if n, err := strconv.Atoi(pageSizeStr); err == nil && n > 0 {
			pageSize = n
			hasPageSize = true
		}
	}

	if pageNumberStr != "" {
		if n, err := strconv.Atoi(pageNumberStr); err == nil && n > 0 {
			pageNumber = n
			hasPageNumber = true
		}
	}

	if !hasPageSize && !hasPageNumber {
		return nil
	}

	if hasPageSize && !hasPageNumber {
		pageNumber = DefaultPageNum
	}

	if !hasPageSize && hasPageNumber {
		pageSize = DefaultPageSize
	}

	return &Req{
		PageSize:   pageSize,
		PageNumber: pageNumber,
	}
}

func GetPage[T any](req *Req, data []T) ([]T, *Resp) {
	if req == nil || req.PageSize <= 0 {
		return nil, &Resp{Page: 0, PageSize: 0, TotalItems: 0}
	}

	if req.PageNumber < 1 {
		req.PageNumber = 1
	}

	total := len(data)

	offset := req.GetOffset()
	if offset >= total {
		return []T{}, &Resp{Page: req.PageNumber, PageSize: req.PageSize, TotalItems: total}
	}

	end := offset + req.PageSize
	if end > total {
		end = total
	}

	return data[offset:end], &Resp{Page: req.PageNumber, PageSize: req.PageSize, TotalItems: total}
}
