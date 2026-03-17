package pagination

import (
	"net/http"
	"strconv"
)

type Params struct {
	Page  int
	Limit int
}

type Meta struct {
	Total int `json:"total"`
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

func Parse(r *http.Request) Params {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	return Params{Page: page, Limit: limit}
}

func (p Params) Offset() int {
	return (p.Page - 1) * p.Limit
}
