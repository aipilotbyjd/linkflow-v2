package dto

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/repositories"
)

// Note requests

type CreateNoteRequest struct {
	ResourceID   string       `json:"resource_id" validate:"required,uuid"`
	ResourceName string       `json:"resource_name" validate:"required,min=1,max=50"`
	Content      string       `json:"content" validate:"required,min=1"`
	Position     *models.JSON `json:"position,omitempty"`
	Size         *models.JSON `json:"size,omitempty"`
	Color        *string      `json:"color,omitempty" validate:"omitempty,max=20"`
}

type UpdateNoteRequest struct {
	Content  *string      `json:"content,omitempty" validate:"omitempty,min=1"`
	Position *models.JSON `json:"position,omitempty"`
	Size     *models.JSON `json:"size,omitempty"`
	Color    *string      `json:"color,omitempty" validate:"omitempty,max=20"`
}

// Note responses

type NoteResponse struct {
	ID           string      `json:"id"`
	WorkspaceID  string      `json:"workspace_id"`
	ResourceID   string      `json:"resource_id"`
	ResourceName string      `json:"resource_name"`
	CreatedBy    string      `json:"created_by"`
	Content      string      `json:"content"`
	Position     models.JSON `json:"position"`
	Size         models.JSON `json:"size"`
	Color        string      `json:"color"`
	CreatedAt    int64       `json:"created_at"`
	UpdatedAt    int64       `json:"updated_at"`
}

// BuildNoteResponse converts a Note model to NoteResponse
func BuildNoteResponse(note *models.Note) NoteResponse {
	return NoteResponse{
		ID:           note.ID.String(),
		WorkspaceID:  note.WorkspaceID.String(),
		ResourceID:   note.ResourceID.String(),
		ResourceName: note.ResourceName,
		CreatedBy:    note.CreatedBy.String(),
		Content:      note.Content,
		Position:     note.Position,
		Size:         note.Size,
		Color:        note.Color,
		CreatedAt:    note.CreatedAt.Unix(),
		UpdatedAt:    note.UpdatedAt.Unix(),
	}
}

// Note filters

type NoteFilters struct {
	ResourceID   *uuid.UUID
	ResourceName *string
	Color        *string
	Search       *string
	SortBy       string
	Order        string
}

// ParseNoteFilters parses note filters from request query parameters
func ParseNoteFilters(r *http.Request) *NoteFilters {
	filters := &NoteFilters{
		SortBy: "created_at",
		Order:  "desc",
	}

	if resourceID := r.URL.Query().Get("resource_id"); resourceID != "" {
		if id, err := uuid.Parse(resourceID); err == nil {
			filters.ResourceID = &id
		}
	}

	if resourceName := r.URL.Query().Get("resource_name"); resourceName != "" {
		filters.ResourceName = &resourceName
	}

	if color := r.URL.Query().Get("color"); color != "" {
		filters.Color = &color
	}

	if search := r.URL.Query().Get("search"); search != "" {
		filters.Search = &search
	}

	if sortBy := r.URL.Query().Get("sort_by"); sortBy != "" {
		filters.SortBy = sortBy
	}

	if order := r.URL.Query().Get("order"); order == "asc" || order == "desc" {
		filters.Order = order
	}

	return filters
}

// ToRepoFilter converts DTO filters to repository filters
func (f *NoteFilters) ToRepoFilter() *repositories.NoteFilter {
	return &repositories.NoteFilter{
		ResourceID:   f.ResourceID,
		ResourceName: f.ResourceName,
		Color:        f.Color,
		Search:       f.Search,
		SortBy:       f.SortBy,
		Order:        f.Order,
	}
}

// ToQueryString builds query string from filters
func (f *NoteFilters) ToQueryString() string {
	qs := ""
	if f.ResourceID != nil {
		qs += "&resource_id=" + f.ResourceID.String()
	}
	if f.ResourceName != nil && *f.ResourceName != "" {
		qs += "&resource_name=" + *f.ResourceName
	}
	if f.Color != nil && *f.Color != "" {
		qs += "&color=" + *f.Color
	}
	if f.Search != nil && *f.Search != "" {
		qs += "&search=" + *f.Search
	}
	if f.SortBy != "" && f.SortBy != "created_at" {
		qs += "&sort_by=" + f.SortBy
	}
	if f.Order != "" && f.Order != "desc" {
		qs += "&order=" + f.Order
	}
	return qs
}
