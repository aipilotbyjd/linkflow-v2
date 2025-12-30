package dto

import (
	"net/http"
	"strconv"

	"github.com/linkflow-ai/linkflow/internal/domain/repositories"
)

// Pagination holds parsed pagination parameters
type Pagination struct {
	Page    int
	PerPage int
	Opts    *repositories.ListOptions
}

// ParsePagination extracts pagination parameters from request query.
func ParsePagination(r *http.Request) *Pagination {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))

	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}

	return &Pagination{
		Page:    page,
		PerPage: perPage,
		Opts:    repositories.NewListOptions(page, perPage),
	}
}

// NewMeta creates a Meta response from pagination and total count
func (p *Pagination) NewMeta(total int64) *Meta {
	totalPages := int(total) / p.PerPage
	if int(total)%p.PerPage > 0 {
		totalPages++
	}
	return &Meta{
		Page:       p.Page,
		PerPage:    p.PerPage,
		Total:      total,
		TotalPages: totalPages,
	}
}

// ParseBool parses a boolean from query parameter
func ParseBool(r *http.Request, param string) bool {
	val := r.URL.Query().Get(param)
	return val == "true" || val == "1"
}

// QueryString returns query parameter or empty string
func QueryString(r *http.Request, param string) string {
	return r.URL.Query().Get(param)
}
