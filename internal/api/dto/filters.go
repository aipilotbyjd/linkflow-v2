package dto

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// WorkflowFilters contains filter parameters for workflow listing
type WorkflowFilters struct {
	Status        *string    `query:"status"`         // active, inactive, draft
	Search        *string    `query:"search"`         // Search in name/description
	Tags          []string   `query:"tags"`           // Filter by tags (comma-separated)
	CreatedAfter  *time.Time `query:"created_after"`  // Filter by creation date
	CreatedBefore *time.Time `query:"created_before"` // Filter by creation date
	UpdatedAfter  *time.Time `query:"updated_after"`  // Filter by update date
	UpdatedBefore *time.Time `query:"updated_before"` // Filter by update date
	SortBy        string     `query:"sort_by"`        // name, created_at, updated_at, execution_count
	Order         string     `query:"order"`          // asc, desc
}

// ParseWorkflowFilters parses workflow filter parameters from HTTP request
func ParseWorkflowFilters(r *http.Request) *WorkflowFilters {
	f := &WorkflowFilters{}
	q := r.URL.Query()

	// Status filter
	if status := q.Get("status"); status != "" {
		if isValidWorkflowStatus(status) {
			f.Status = &status
		}
	}

	// Search filter
	if search := q.Get("search"); search != "" {
		f.Search = &search
	}

	// Tags filter (comma-separated)
	if tags := q.Get("tags"); tags != "" {
		f.Tags = strings.Split(tags, ",")
		// Trim whitespace from each tag
		for i := range f.Tags {
			f.Tags[i] = strings.TrimSpace(f.Tags[i])
		}
	}

	// Date filters
	if createdAfter := q.Get("created_after"); createdAfter != "" {
		if t, err := parseDateTime(createdAfter); err == nil {
			f.CreatedAfter = &t
		}
	}
	if createdBefore := q.Get("created_before"); createdBefore != "" {
		if t, err := parseDateTime(createdBefore); err == nil {
			f.CreatedBefore = &t
		}
	}
	if updatedAfter := q.Get("updated_after"); updatedAfter != "" {
		if t, err := parseDateTime(updatedAfter); err == nil {
			f.UpdatedAfter = &t
		}
	}
	if updatedBefore := q.Get("updated_before"); updatedBefore != "" {
		if t, err := parseDateTime(updatedBefore); err == nil {
			f.UpdatedBefore = &t
		}
	}

	// Sorting
	if sortBy := q.Get("sort_by"); sortBy != "" && isValidWorkflowSortField(sortBy) {
		f.SortBy = sortBy
	} else {
		f.SortBy = "created_at" // default
	}
	if order := q.Get("order"); order != "" && isValidOrder(order) {
		f.Order = strings.ToLower(order)
	} else {
		f.Order = "desc" // default
	}

	return f
}

// HasFilters returns true if any filter is applied
func (f *WorkflowFilters) HasFilters() bool {
	return f.Status != nil || f.Search != nil || len(f.Tags) > 0 ||
		f.CreatedAfter != nil || f.CreatedBefore != nil ||
		f.UpdatedAfter != nil || f.UpdatedBefore != nil
}

// ToQueryString converts filters to query string for pagination links
func (f *WorkflowFilters) ToQueryString() string {
	var parts []string

	if f.Status != nil {
		parts = append(parts, "status="+*f.Status)
	}
	if f.Search != nil {
		parts = append(parts, "search="+*f.Search)
	}
	if len(f.Tags) > 0 {
		parts = append(parts, "tags="+strings.Join(f.Tags, ","))
	}
	if f.CreatedAfter != nil {
		parts = append(parts, "created_after="+f.CreatedAfter.Format(time.RFC3339))
	}
	if f.CreatedBefore != nil {
		parts = append(parts, "created_before="+f.CreatedBefore.Format(time.RFC3339))
	}
	if f.UpdatedAfter != nil {
		parts = append(parts, "updated_after="+f.UpdatedAfter.Format(time.RFC3339))
	}
	if f.UpdatedBefore != nil {
		parts = append(parts, "updated_before="+f.UpdatedBefore.Format(time.RFC3339))
	}
	if f.SortBy != "" && f.SortBy != "created_at" {
		parts = append(parts, "sort_by="+f.SortBy)
	}
	if f.Order != "" && f.Order != "desc" {
		parts = append(parts, "order="+f.Order)
	}

	if len(parts) == 0 {
		return ""
	}
	return "&" + strings.Join(parts, "&")
}

// CredentialFilters contains filter parameters for credential listing
type CredentialFilters struct {
	Type   *string `query:"type"`   // Filter by credential type
	Search *string `query:"search"` // Search in name/description
	SortBy string  `query:"sort_by"`
	Order  string  `query:"order"`
}

// ParseCredentialFilters parses credential filter parameters from HTTP request
func ParseCredentialFilters(r *http.Request) *CredentialFilters {
	f := &CredentialFilters{}
	q := r.URL.Query()

	if credType := q.Get("type"); credType != "" {
		f.Type = &credType
	}
	if search := q.Get("search"); search != "" {
		f.Search = &search
	}
	if sortBy := q.Get("sort_by"); sortBy != "" && isValidCredentialSortField(sortBy) {
		f.SortBy = sortBy
	} else {
		f.SortBy = "created_at"
	}
	if order := q.Get("order"); order != "" && isValidOrder(order) {
		f.Order = strings.ToLower(order)
	} else {
		f.Order = "desc"
	}

	return f
}

// HasFilters returns true if any filter is applied
func (f *CredentialFilters) HasFilters() bool {
	return f.Type != nil || f.Search != nil
}

// ToQueryString converts filters to query string for pagination links
func (f *CredentialFilters) ToQueryString() string {
	var parts []string

	if f.Type != nil {
		parts = append(parts, "type="+*f.Type)
	}
	if f.Search != nil {
		parts = append(parts, "search="+*f.Search)
	}
	if f.SortBy != "" && f.SortBy != "created_at" {
		parts = append(parts, "sort_by="+f.SortBy)
	}
	if f.Order != "" && f.Order != "desc" {
		parts = append(parts, "order="+f.Order)
	}

	if len(parts) == 0 {
		return ""
	}
	return "&" + strings.Join(parts, "&")
}

// ScheduleFilters contains filter parameters for schedule listing
type ScheduleFilters struct {
	IsActive   *bool   `query:"is_active"`   // Filter by active status
	WorkflowID *string `query:"workflow_id"` // Filter by workflow
	Search     *string `query:"search"`      // Search in name
	SortBy     string  `query:"sort_by"`
	Order      string  `query:"order"`
}

// ParseScheduleFilters parses schedule filter parameters from HTTP request
func ParseScheduleFilters(r *http.Request) *ScheduleFilters {
	f := &ScheduleFilters{}
	q := r.URL.Query()

	if isActive := q.Get("is_active"); isActive != "" {
		val := isActive == "true" || isActive == "1"
		f.IsActive = &val
	}
	if workflowID := q.Get("workflow_id"); workflowID != "" {
		f.WorkflowID = &workflowID
	}
	if search := q.Get("search"); search != "" {
		f.Search = &search
	}
	if sortBy := q.Get("sort_by"); sortBy != "" && isValidScheduleSortField(sortBy) {
		f.SortBy = sortBy
	} else {
		f.SortBy = "created_at"
	}
	if order := q.Get("order"); order != "" && isValidOrder(order) {
		f.Order = strings.ToLower(order)
	} else {
		f.Order = "desc"
	}

	return f
}

// HasFilters returns true if any filter is applied
func (f *ScheduleFilters) HasFilters() bool {
	return f.IsActive != nil || f.WorkflowID != nil || f.Search != nil
}

// ToQueryString converts filters to query string for pagination links
func (f *ScheduleFilters) ToQueryString() string {
	var parts []string

	if f.IsActive != nil {
		if *f.IsActive {
			parts = append(parts, "is_active=true")
		} else {
			parts = append(parts, "is_active=false")
		}
	}
	if f.WorkflowID != nil {
		parts = append(parts, "workflow_id="+*f.WorkflowID)
	}
	if f.Search != nil {
		parts = append(parts, "search="+*f.Search)
	}
	if f.SortBy != "" && f.SortBy != "created_at" {
		parts = append(parts, "sort_by="+f.SortBy)
	}
	if f.Order != "" && f.Order != "desc" {
		parts = append(parts, "order="+f.Order)
	}

	if len(parts) == 0 {
		return ""
	}
	return "&" + strings.Join(parts, "&")
}

// WebhookFilters contains filter parameters for webhook endpoint listing
type WebhookFilters struct {
	IsActive   *bool   `query:"is_active"`   // Filter by active status
	WorkflowID *string `query:"workflow_id"` // Filter by workflow
	Method     *string `query:"method"`      // Filter by HTTP method
	Search     *string `query:"search"`      // Search in path
	SortBy     string  `query:"sort_by"`
	Order      string  `query:"order"`
}

// ParseWebhookFilters parses webhook filter parameters from HTTP request
func ParseWebhookFilters(r *http.Request) *WebhookFilters {
	f := &WebhookFilters{}
	q := r.URL.Query()

	if isActive := q.Get("is_active"); isActive != "" {
		val := isActive == "true" || isActive == "1"
		f.IsActive = &val
	}
	if workflowID := q.Get("workflow_id"); workflowID != "" {
		f.WorkflowID = &workflowID
	}
	if method := q.Get("method"); method != "" && isValidHTTPMethod(method) {
		upper := strings.ToUpper(method)
		f.Method = &upper
	}
	if search := q.Get("search"); search != "" {
		f.Search = &search
	}
	if sortBy := q.Get("sort_by"); sortBy != "" && isValidWebhookSortField(sortBy) {
		f.SortBy = sortBy
	} else {
		f.SortBy = "created_at"
	}
	if order := q.Get("order"); order != "" && isValidOrder(order) {
		f.Order = strings.ToLower(order)
	} else {
		f.Order = "desc"
	}

	return f
}

// HasFilters returns true if any filter is applied
func (f *WebhookFilters) HasFilters() bool {
	return f.IsActive != nil || f.WorkflowID != nil || f.Method != nil || f.Search != nil
}

// ToQueryString converts filters to query string for pagination links
func (f *WebhookFilters) ToQueryString() string {
	var parts []string

	if f.IsActive != nil {
		if *f.IsActive {
			parts = append(parts, "is_active=true")
		} else {
			parts = append(parts, "is_active=false")
		}
	}
	if f.WorkflowID != nil {
		parts = append(parts, "workflow_id="+*f.WorkflowID)
	}
	if f.Method != nil {
		parts = append(parts, "method="+*f.Method)
	}
	if f.Search != nil {
		parts = append(parts, "search="+*f.Search)
	}
	if f.SortBy != "" && f.SortBy != "created_at" {
		parts = append(parts, "sort_by="+f.SortBy)
	}
	if f.Order != "" && f.Order != "desc" {
		parts = append(parts, "order="+f.Order)
	}

	if len(parts) == 0 {
		return ""
	}
	return "&" + strings.Join(parts, "&")
}

// ExecutionFilters contains filter parameters for execution listing
type ExecutionFilters struct {
	Status      *string    `query:"status"`       // pending, running, completed, failed, cancelled
	WorkflowID  *string    `query:"workflow_id"`  // Filter by workflow
	TriggerType *string    `query:"trigger_type"` // manual, webhook, schedule
	StartDate   *time.Time `query:"start_date"`   // Filter by start date
	EndDate     *time.Time `query:"end_date"`     // Filter by end date
	Search      *string    `query:"search"`       // Search in error message
	SortBy      string     `query:"sort_by"`
	Order       string     `query:"order"`
}

// ParseExecutionFilters parses execution filter parameters from HTTP request
func ParseExecutionFilters(r *http.Request) *ExecutionFilters {
	f := &ExecutionFilters{}
	q := r.URL.Query()

	if status := q.Get("status"); status != "" && isValidExecutionStatus(status) {
		f.Status = &status
	}
	if workflowID := q.Get("workflow_id"); workflowID != "" {
		f.WorkflowID = &workflowID
	}
	if triggerType := q.Get("trigger_type"); triggerType != "" && isValidTriggerType(triggerType) {
		f.TriggerType = &triggerType
	}
	if startDate := q.Get("start_date"); startDate != "" {
		if t, err := parseDateTime(startDate); err == nil {
			f.StartDate = &t
		}
	}
	if endDate := q.Get("end_date"); endDate != "" {
		if t, err := parseDateTime(endDate); err == nil {
			f.EndDate = &t
		}
	}
	if search := q.Get("search"); search != "" {
		f.Search = &search
	}
	if sortBy := q.Get("sort_by"); sortBy != "" && isValidExecutionSortField(sortBy) {
		f.SortBy = sortBy
	} else {
		f.SortBy = "queued_at"
	}
	if order := q.Get("order"); order != "" && isValidOrder(order) {
		f.Order = strings.ToLower(order)
	} else {
		f.Order = "desc"
	}

	return f
}

// HasFilters returns true if any filter is applied
func (f *ExecutionFilters) HasFilters() bool {
	return f.Status != nil || f.WorkflowID != nil || f.TriggerType != nil ||
		f.StartDate != nil || f.EndDate != nil || f.Search != nil
}

// ToQueryString converts filters to query string for pagination links
func (f *ExecutionFilters) ToQueryString() string {
	var parts []string

	if f.Status != nil {
		parts = append(parts, "status="+*f.Status)
	}
	if f.WorkflowID != nil {
		parts = append(parts, "workflow_id="+*f.WorkflowID)
	}
	if f.TriggerType != nil {
		parts = append(parts, "trigger_type="+*f.TriggerType)
	}
	if f.StartDate != nil {
		parts = append(parts, "start_date="+f.StartDate.Format(time.RFC3339))
	}
	if f.EndDate != nil {
		parts = append(parts, "end_date="+f.EndDate.Format(time.RFC3339))
	}
	if f.Search != nil {
		parts = append(parts, "search="+*f.Search)
	}
	if f.SortBy != "" && f.SortBy != "queued_at" {
		parts = append(parts, "sort_by="+f.SortBy)
	}
	if f.Order != "" && f.Order != "desc" {
		parts = append(parts, "order="+f.Order)
	}

	if len(parts) == 0 {
		return ""
	}
	return "&" + strings.Join(parts, "&")
}

// Helper validation functions

func isValidWorkflowStatus(status string) bool {
	validStatuses := map[string]bool{"active": true, "inactive": true, "draft": true, "archived": true}
	return validStatuses[strings.ToLower(status)]
}

func isValidExecutionStatus(status string) bool {
	validStatuses := map[string]bool{
		"queued": true, "running": true, "completed": true,
		"failed": true, "cancelled": true, "timeout": true,
	}
	return validStatuses[strings.ToLower(status)]
}

func isValidTriggerType(triggerType string) bool {
	validTypes := map[string]bool{"manual": true, "webhook": true, "schedule": true}
	return validTypes[strings.ToLower(triggerType)]
}

func isValidOrder(order string) bool {
	return strings.ToLower(order) == "asc" || strings.ToLower(order) == "desc"
}

func isValidWorkflowSortField(field string) bool {
	validFields := map[string]bool{
		"name": true, "created_at": true, "updated_at": true,
		"execution_count": true, "last_executed_at": true,
	}
	return validFields[strings.ToLower(field)]
}

func isValidCredentialSortField(field string) bool {
	validFields := map[string]bool{
		"name": true, "created_at": true, "type": true, "last_used_at": true,
	}
	return validFields[strings.ToLower(field)]
}

func isValidScheduleSortField(field string) bool {
	validFields := map[string]bool{
		"created_at": true, "next_run_at": true, "last_run_at": true,
	}
	return validFields[strings.ToLower(field)]
}

func isValidWebhookSortField(field string) bool {
	validFields := map[string]bool{
		"created_at": true, "path": true, "last_triggered_at": true,
	}
	return validFields[strings.ToLower(field)]
}

func isValidExecutionSortField(field string) bool {
	validFields := map[string]bool{
		"queued_at": true, "started_at": true, "completed_at": true, "duration": true,
	}
	return validFields[strings.ToLower(field)]
}

func isValidHTTPMethod(method string) bool {
	validMethods := map[string]bool{
		"GET": true, "POST": true, "PUT": true, "PATCH": true, "DELETE": true,
	}
	return validMethods[strings.ToUpper(method)]
}

func parseDateTime(value string) (time.Time, error) {
	// Try RFC3339 first
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t, nil
	}
	// Try date only format
	if t, err := time.Parse("2006-01-02", value); err == nil {
		return t, nil
	}
	// Try datetime without timezone
	if t, err := time.Parse("2006-01-02T15:04:05", value); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("invalid date format: %s", value)
}
