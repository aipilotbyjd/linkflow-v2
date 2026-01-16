package types

import (
	"time"

	"github.com/google/uuid"
)

const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// PageRequest represents offset-based pagination parameters
type PageRequest struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	SortBy   string `json:"sort_by"`
	SortDesc bool   `json:"sort_desc"`
}

func NewPageRequest(page, pageSize int) PageRequest {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}
	return PageRequest{
		Page:     page,
		PageSize: pageSize,
		SortBy:   "created_at",
		SortDesc: true,
	}
}

func (p PageRequest) Offset() int {
	return (p.Page - 1) * p.PageSize
}

func (p PageRequest) Limit() int {
	return p.PageSize
}

// PageResponse represents paginated response metadata
type PageResponse struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
	HasMore    bool  `json:"has_more"`
}

func NewPageResponse(page, pageSize int, totalItems int64) PageResponse {
	totalPages := int(totalItems) / pageSize
	if int(totalItems)%pageSize != 0 {
		totalPages++
	}
	return PageResponse{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: totalItems,
		TotalPages: totalPages,
		HasMore:    page < totalPages,
	}
}

// CursorRequest represents cursor-based pagination parameters
type CursorRequest struct {
	Limit      int        `json:"limit"`
	CursorID   *uuid.UUID `json:"cursor_id,omitempty"`
	CursorTime *time.Time `json:"cursor_time,omitempty"`
}

func NewCursorRequest(limit int, cursorID *uuid.UUID, cursorTime *time.Time) CursorRequest {
	if limit < 1 {
		limit = DefaultPageSize
	}
	if limit > MaxPageSize {
		limit = MaxPageSize
	}
	return CursorRequest{
		Limit:      limit,
		CursorID:   cursorID,
		CursorTime: cursorTime,
	}
}

func (c CursorRequest) HasCursor() bool {
	return c.CursorID != nil && c.CursorTime != nil
}

// CursorResponse represents cursor-based pagination response
type CursorResponse struct {
	NextCursorID   *uuid.UUID `json:"next_cursor_id,omitempty"`
	NextCursorTime *time.Time `json:"next_cursor_time,omitempty"`
	HasMore        bool       `json:"has_more"`
}

// ListOptions represents common listing options for repositories
type ListOptions struct {
	Offset     int
	Limit      int
	OrderBy    string
	Order      string
	CursorID   *uuid.UUID
	CursorTime *time.Time
	UseCursor  bool
}

func NewListOptions(page, perPage int) *ListOptions {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = DefaultPageSize
	}
	if perPage > MaxPageSize {
		perPage = MaxPageSize
	}
	return &ListOptions{
		Offset:  (page - 1) * perPage,
		Limit:   perPage,
		OrderBy: "created_at",
		Order:   "desc",
	}
}

func NewCursorListOptions(limit int, cursorID *uuid.UUID, cursorTime *time.Time) *ListOptions {
	if limit < 1 {
		limit = DefaultPageSize
	}
	if limit > MaxPageSize {
		limit = MaxPageSize
	}
	return &ListOptions{
		Limit:      limit + 1,
		OrderBy:    "created_at",
		Order:      "desc",
		CursorID:   cursorID,
		CursorTime: cursorTime,
		UseCursor:  true,
	}
}
