package dto

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// EnhancedResponse provides frontend-friendly features
type EnhancedResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data"`
	Included  interface{} `json:"included,omitempty"`  // Related resources (avoid N+1)
	Links     *Links      `json:"links,omitempty"`     // HATEOAS navigation
	Actions   []Action    `json:"actions,omitempty"`   // Available actions
	Meta      interface{} `json:"meta,omitempty"`      // Pagination or cursor
	RequestID string      `json:"request_id,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

// Links provides navigation URLs
type Links struct {
	Self  string `json:"self,omitempty"`
	Next  string `json:"next,omitempty"`
	Prev  string `json:"prev,omitempty"`
	First string `json:"first,omitempty"`
	Last  string `json:"last,omitempty"`
}

// Action represents an available action on the resource
type Action struct {
	Name   string `json:"name"`             // e.g., "execute", "activate", "delete"
	Method string `json:"method"`           // HTTP method
	Href   string `json:"href"`             // Action URL
	Label  string `json:"label,omitempty"`  // Human-readable label
}

// IncludeBuilder helps build included relationships
type IncludeBuilder struct {
	data map[string]interface{}
}

func NewIncludeBuilder() *IncludeBuilder {
	return &IncludeBuilder{data: make(map[string]interface{})}
}

func (b *IncludeBuilder) Add(key string, value interface{}) *IncludeBuilder {
	b.data[key] = value
	return b
}

func (b *IncludeBuilder) Build() map[string]interface{} {
	if len(b.data) == 0 {
		return nil
	}
	return b.data
}

// ResponseBuilder provides fluent API for building responses
type ResponseBuilder struct {
	status   int
	data     interface{}
	included interface{}
	links    *Links
	actions  []Action
	meta     interface{}
}

func NewResponse(data interface{}) *ResponseBuilder {
	return &ResponseBuilder{
		status: http.StatusOK,
		data:   data,
	}
}

func (b *ResponseBuilder) Status(status int) *ResponseBuilder {
	b.status = status
	return b
}

func (b *ResponseBuilder) WithIncluded(included interface{}) *ResponseBuilder {
	b.included = included
	return b
}

func (b *ResponseBuilder) WithLinks(links *Links) *ResponseBuilder {
	b.links = links
	return b
}

func (b *ResponseBuilder) WithActions(actions ...Action) *ResponseBuilder {
	b.actions = actions
	return b
}

func (b *ResponseBuilder) WithMeta(meta interface{}) *ResponseBuilder {
	b.meta = meta
	return b
}

func (b *ResponseBuilder) WithPagination(pg *Pagination, total int64) *ResponseBuilder {
	b.meta = pg.NewMeta(total)
	return b
}

func (b *ResponseBuilder) WithCursor(cursor *CursorMeta) *ResponseBuilder {
	b.meta = cursor
	return b
}

func (b *ResponseBuilder) Send(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(b.status)

	response := EnhancedResponse{
		Success:   b.status >= 200 && b.status < 300,
		Data:      b.data,
		Included:  b.included,
		Links:     b.links,
		Actions:   b.actions,
		Meta:      b.meta,
		RequestID: getRequestID(w),
		Timestamp: time.Now().Unix(),
	}

	_ = json.NewEncoder(w).Encode(response)
}

// Common action builders

func ExecuteAction(workspaceID, workflowID string) Action {
	return Action{
		Name:   "execute",
		Method: "POST",
		Href:   "/api/v1/workspaces/" + workspaceID + "/workflows/" + workflowID + "/execute",
		Label:  "Execute Workflow",
	}
}

func ActivateAction(workspaceID, workflowID string) Action {
	return Action{
		Name:   "activate",
		Method: "POST",
		Href:   "/api/v1/workspaces/" + workspaceID + "/workflows/" + workflowID + "/activate",
		Label:  "Activate Workflow",
	}
}

func DeactivateAction(workspaceID, workflowID string) Action {
	return Action{
		Name:   "deactivate",
		Method: "POST",
		Href:   "/api/v1/workspaces/" + workspaceID + "/workflows/" + workflowID + "/deactivate",
		Label:  "Deactivate Workflow",
	}
}

func DeleteAction(href string) Action {
	return Action{
		Name:   "delete",
		Method: "DELETE",
		Href:   href,
		Label:  "Delete",
	}
}

func RetryAction(workspaceID, executionID string) Action {
	return Action{
		Name:   "retry",
		Method: "POST",
		Href:   "/api/v1/workspaces/" + workspaceID + "/executions/" + executionID + "/retry",
		Label:  "Retry Execution",
	}
}

func CancelAction(workspaceID, executionID string) Action {
	return Action{
		Name:   "cancel",
		Method: "POST",
		Href:   "/api/v1/workspaces/" + workspaceID + "/executions/" + executionID + "/cancel",
		Label:  "Cancel Execution",
	}
}

// Field selection helpers

// SelectFields filters response fields based on query parameter
// Usage: ?fields=id,name,status
func SelectFields(r *http.Request, data interface{}) interface{} {
	fieldsParam := r.URL.Query().Get("fields")
	if fieldsParam == "" {
		return data
	}

	fields := strings.Split(fieldsParam, ",")
	return filterFields(data, fields)
}

func filterFields(data interface{}, fields []string) interface{} {
	// Convert to map for filtering
	b, err := json.Marshal(data)
	if err != nil {
		return data
	}

	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		// Try as slice
		var slice []map[string]interface{}
		if err := json.Unmarshal(b, &slice); err != nil {
			return data
		}
		return filterSliceFields(slice, fields)
	}

	return filterMapFields(m, fields)
}

func filterMapFields(m map[string]interface{}, fields []string) map[string]interface{} {
	fieldSet := make(map[string]bool)
	for _, f := range fields {
		fieldSet[strings.TrimSpace(f)] = true
	}

	result := make(map[string]interface{})
	for k, v := range m {
		if fieldSet[k] {
			result[k] = v
		}
	}
	return result
}

func filterSliceFields(slice []map[string]interface{}, fields []string) []map[string]interface{} {
	result := make([]map[string]interface{}, len(slice))
	for i, m := range slice {
		result[i] = filterMapFields(m, fields)
	}
	return result
}

// Expand/Include helpers

// ParseIncludes extracts include parameter from request
// Usage: ?include=workflow,user,nodes
func ParseIncludes(r *http.Request) []string {
	includeParam := r.URL.Query().Get("include")
	if includeParam == "" {
		return nil
	}
	includes := strings.Split(includeParam, ",")
	for i := range includes {
		includes[i] = strings.TrimSpace(includes[i])
	}
	return includes
}

// HasInclude checks if a specific include is requested
func HasInclude(includes []string, name string) bool {
	for _, inc := range includes {
		if inc == name {
			return true
		}
	}
	return false
}
