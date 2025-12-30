package dto

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// CursorData holds the encoded cursor information
type CursorData struct {
	ID        string    `json:"id,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	SortValue string    `json:"sort_value,omitempty"`
}

// CursorPagination holds cursor-based pagination parameters
type CursorPagination struct {
	Cursor   string
	Limit    int
	SortBy   string
	SortDesc bool
	Decoded  *CursorData
}

// CursorMeta contains cursor-based pagination metadata
type CursorMeta struct {
	NextCursor string `json:"next_cursor,omitempty"`
	PrevCursor string `json:"prev_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
	Limit      int    `json:"limit"`
}

// ParseCursorPagination extracts cursor pagination parameters from request.
// Query params: cursor, limit, sort_by, sort_desc
func ParseCursorPagination(r *http.Request) *CursorPagination {
	cursor := r.URL.Query().Get("cursor")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	sortBy := r.URL.Query().Get("sort_by")
	sortDesc := r.URL.Query().Get("sort_desc") == "true"

	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	cp := &CursorPagination{
		Cursor:   cursor,
		Limit:    limit,
		SortBy:   sortBy,
		SortDesc: sortDesc,
	}

	// Decode cursor if present
	if cursor != "" {
		cp.Decoded = DecodeCursor(cursor)
	}

	return cp
}

// EncodeCursor creates a base64-encoded cursor from cursor data
func EncodeCursor(data *CursorData) string {
	if data == nil {
		return ""
	}
	b, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

// DecodeCursor decodes a base64 cursor string to cursor data
func DecodeCursor(cursor string) *CursorData {
	if cursor == "" {
		return nil
	}
	b, err := base64.URLEncoding.DecodeString(cursor)
	if err != nil {
		return nil
	}
	var data CursorData
	if err := json.Unmarshal(b, &data); err != nil {
		return nil
	}
	return &data
}

// NewCursorFromID creates a cursor from an ID
func NewCursorFromID(id uuid.UUID) string {
	return EncodeCursor(&CursorData{ID: id.String()})
}

// NewCursorFromTime creates a cursor from a timestamp and ID
func NewCursorFromTime(t time.Time, id uuid.UUID) string {
	return EncodeCursor(&CursorData{
		ID:        id.String(),
		CreatedAt: t,
	})
}

// NewCursorMeta creates cursor metadata from results
// lastID is the ID of the last item in the current page
// hasMore indicates if there are more results
func NewCursorMeta(lastID uuid.UUID, lastTime time.Time, hasMore bool, limit int) *CursorMeta {
	meta := &CursorMeta{
		HasMore: hasMore,
		Limit:   limit,
	}

	if hasMore && lastID != uuid.Nil {
		meta.NextCursor = NewCursorFromTime(lastTime, lastID)
	}

	return meta
}

// JSONWithCursor sends a JSON response with cursor pagination metadata
func JSONWithCursor(w http.ResponseWriter, status int, data interface{}, meta *CursorMeta) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := struct {
		Success   bool        `json:"success"`
		Data      interface{} `json:"data"`
		Cursor    *CursorMeta `json:"cursor,omitempty"`
		RequestID string      `json:"request_id,omitempty"`
		Timestamp int64       `json:"timestamp"`
	}{
		Success:   status >= 200 && status < 300,
		Data:      data,
		Cursor:    meta,
		RequestID: getRequestID(w),
		Timestamp: time.Now().Unix(),
	}

	_ = json.NewEncoder(w).Encode(response)
}
