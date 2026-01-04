package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/repositories"
	"gorm.io/gorm"
)

type DashboardService struct {
	db             *gorm.DB
	workflowRepo   *repositories.WorkflowRepository
	executionRepo  *repositories.ExecutionRepository
	credentialRepo *repositories.CredentialRepository
	scheduleRepo   *repositories.ScheduleRepository
}

func NewDashboardService(
	db *gorm.DB,
	workflowRepo *repositories.WorkflowRepository,
	executionRepo *repositories.ExecutionRepository,
	credentialRepo *repositories.CredentialRepository,
	scheduleRepo *repositories.ScheduleRepository,
) *DashboardService {
	return &DashboardService{
		db:             db,
		workflowRepo:   workflowRepo,
		executionRepo:  executionRepo,
		credentialRepo: credentialRepo,
		scheduleRepo:   scheduleRepo,
	}
}

// DashboardSummary contains all dashboard data
type DashboardSummary struct {
	Summary            SummaryStats          `json:"summary"`
	RecentExecutions   []ExecutionSummary    `json:"recent_executions"`
	TopWorkflows       []WorkflowStats       `json:"top_workflows"`
	RecentFailures     []FailureSummary      `json:"recent_failures"`
	ExecutionsByDay    []DailyExecutions     `json:"executions_by_day"`
	ExecutionsByHour   []HourlyExecutions    `json:"executions_by_hour"`
	UpcomingSchedules  []ScheduleSummary     `json:"upcoming_schedules"`
	ExecutionsByStatus []StatusCount         `json:"executions_by_status"`
	TriggerTypeStats   []TriggerTypeCount    `json:"trigger_type_stats"`
}

type SummaryStats struct {
	TotalWorkflows         int     `json:"total_workflows"`
	ActiveWorkflows        int     `json:"active_workflows"`
	InactiveWorkflows      int     `json:"inactive_workflows"`
	DraftWorkflows         int     `json:"draft_workflows"`
	TotalExecutionsToday   int64   `json:"total_executions_today"`
	TotalExecutionsWeek    int64   `json:"total_executions_week"`
	TotalExecutionsMonth   int64   `json:"total_executions_month"`
	SuccessRate            float64 `json:"success_rate"`
	AvgDurationMs          int64   `json:"avg_duration_ms"`
	TotalCredentials       int64   `json:"total_credentials"`
	TotalSchedules         int64   `json:"total_schedules"`
	ActiveSchedules        int64   `json:"active_schedules"`
	RunningExecutions      int64   `json:"running_executions"`
	QueuedExecutions       int64   `json:"queued_executions"`
}

type ExecutionSummary struct {
	ID           string  `json:"id"`
	WorkflowID   string  `json:"workflow_id"`
	WorkflowName string  `json:"workflow_name"`
	Status       string  `json:"status"`
	TriggerType  string  `json:"trigger_type"`
	DurationMs   *int64  `json:"duration_ms,omitempty"`
	StartedAt    *int64  `json:"started_at,omitempty"`
	CompletedAt  *int64  `json:"completed_at,omitempty"`
	CreatedAt    int64   `json:"created_at"`
}

type WorkflowStats struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Status         string  `json:"status"`
	ExecutionCount int64   `json:"execution_count"`
	SuccessCount   int64   `json:"success_count"`
	FailedCount    int64   `json:"failed_count"`
	SuccessRate    float64 `json:"success_rate"`
	AvgDurationMs  int64   `json:"avg_duration_ms"`
	LastExecutedAt *int64  `json:"last_executed_at,omitempty"`
}

type FailureSummary struct {
	ID           string `json:"id"`
	WorkflowID   string `json:"workflow_id"`
	WorkflowName string `json:"workflow_name"`
	ErrorMessage string `json:"error_message"`
	ErrorNodeID  string `json:"error_node_id,omitempty"`
	FailedAt     int64  `json:"failed_at"`
}

type DailyExecutions struct {
	Date    string `json:"date"`
	Total   int64  `json:"total"`
	Success int64  `json:"success"`
	Failed  int64  `json:"failed"`
}

type HourlyExecutions struct {
	Hour  int   `json:"hour"`
	Count int64 `json:"count"`
}

type ScheduleSummary struct {
	ID             string `json:"id"`
	WorkflowID     string `json:"workflow_id"`
	WorkflowName   string `json:"workflow_name"`
	CronExpression string `json:"cron_expression"`
	Timezone       string `json:"timezone"`
	NextRunAt      *int64 `json:"next_run_at,omitempty"`
	IsActive       bool   `json:"is_active"`
}

type StatusCount struct {
	Status string `json:"status"`
	Count  int64  `json:"count"`
}

type TriggerTypeCount struct {
	TriggerType string `json:"trigger_type"`
	Count       int64  `json:"count"`
}

// QuickStats for sidebar/header
type QuickStats struct {
	Workflows    WorkflowQuickStats    `json:"workflows"`
	Executions   ExecutionQuickStats   `json:"executions"`
	Credentials  CredentialQuickStats  `json:"credentials"`
	Schedules    ScheduleQuickStats    `json:"schedules"`
}

type WorkflowQuickStats struct {
	Total  int64 `json:"total"`
	Active int64 `json:"active"`
}

type ExecutionQuickStats struct {
	Running int64 `json:"running"`
	Queued  int64 `json:"queued"`
	Today   int64 `json:"today"`
}

type CredentialQuickStats struct {
	Total        int64 `json:"total"`
	ExpiringSoon int64 `json:"expiring_soon"`
}

type ScheduleQuickStats struct {
	Total  int64 `json:"total"`
	Active int64 `json:"active"`
}

// GetDashboard returns full dashboard data
func (s *DashboardService) GetDashboard(ctx context.Context, workspaceID uuid.UUID, period string) (*DashboardSummary, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	startOfWeek := startOfDay.AddDate(0, 0, -7)
	startOfMonth := startOfDay.AddDate(0, -1, 0)

	// Determine chart period
	var chartStart time.Time
	switch period {
	case "7d":
		chartStart = startOfWeek
	case "30d":
		chartStart = startOfMonth
	case "90d":
		chartStart = startOfDay.AddDate(0, -3, 0)
	default:
		chartStart = startOfWeek
	}

	dashboard := &DashboardSummary{}

	// Get summary stats
	summary, err := s.getSummaryStats(ctx, workspaceID, startOfDay, startOfWeek, startOfMonth)
	if err != nil {
		return nil, err
	}
	dashboard.Summary = *summary

	// Get recent executions (last 10)
	recentExecs, err := s.getRecentExecutions(ctx, workspaceID, 10)
	if err != nil {
		return nil, err
	}
	dashboard.RecentExecutions = recentExecs

	// Get top workflows (by execution count in period)
	topWorkflows, err := s.getTopWorkflows(ctx, workspaceID, chartStart, 5)
	if err != nil {
		return nil, err
	}
	dashboard.TopWorkflows = topWorkflows

	// Get recent failures (last 10)
	recentFailures, err := s.getRecentFailures(ctx, workspaceID, 10)
	if err != nil {
		return nil, err
	}
	dashboard.RecentFailures = recentFailures

	// Get executions by day
	execsByDay, err := s.getExecutionsByDay(ctx, workspaceID, chartStart)
	if err != nil {
		return nil, err
	}
	dashboard.ExecutionsByDay = execsByDay

	// Get executions by hour (last 24h)
	execsByHour, err := s.getExecutionsByHour(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	dashboard.ExecutionsByHour = execsByHour

	// Get upcoming schedules
	upcomingSchedules, err := s.getUpcomingSchedules(ctx, workspaceID, 5)
	if err != nil {
		return nil, err
	}
	dashboard.UpcomingSchedules = upcomingSchedules

	// Get executions by status (for pie chart)
	statusCounts, err := s.getExecutionsByStatus(ctx, workspaceID, chartStart)
	if err != nil {
		return nil, err
	}
	dashboard.ExecutionsByStatus = statusCounts

	// Get trigger type stats
	triggerStats, err := s.getTriggerTypeStats(ctx, workspaceID, chartStart)
	if err != nil {
		return nil, err
	}
	dashboard.TriggerTypeStats = triggerStats

	return dashboard, nil
}

// GetQuickStats returns lightweight stats for sidebar
func (s *DashboardService) GetQuickStats(ctx context.Context, workspaceID uuid.UUID) (*QuickStats, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	stats := &QuickStats{}

	// Workflow stats
	var workflowCounts []struct {
		Status string `gorm:"column:status"`
		Count  int64  `gorm:"column:count"`
	}
	s.db.WithContext(ctx).
		Model(&models.Workflow{}).
		Select("status, COUNT(*) as count").
		Where("workspace_id = ?", workspaceID).
		Group("status").
		Find(&workflowCounts)

	for _, wc := range workflowCounts {
		stats.Workflows.Total += wc.Count
		if wc.Status == "active" {
			stats.Workflows.Active = wc.Count
		}
	}

	// Execution stats
	var execCounts []struct {
		Status string `gorm:"column:status"`
		Count  int64  `gorm:"column:count"`
	}
	s.db.WithContext(ctx).
		Model(&models.Execution{}).
		Select("status, COUNT(*) as count").
		Where("workspace_id = ? AND created_at >= ?", workspaceID, startOfDay).
		Group("status").
		Find(&execCounts)

	for _, ec := range execCounts {
		stats.Executions.Today += ec.Count
		if ec.Status == "running" {
			stats.Executions.Running = ec.Count
		} else if ec.Status == "queued" {
			stats.Executions.Queued = ec.Count
		}
	}

	// Credential stats
	s.db.WithContext(ctx).
		Model(&models.Credential{}).
		Where("workspace_id = ?", workspaceID).
		Count(&stats.Credentials.Total)

	// Schedule stats
	var scheduleCounts []struct {
		IsActive bool  `gorm:"column:is_active"`
		Count    int64 `gorm:"column:count"`
	}
	s.db.WithContext(ctx).
		Model(&models.Schedule{}).
		Select("is_active, COUNT(*) as count").
		Where("workspace_id = ?", workspaceID).
		Group("is_active").
		Find(&scheduleCounts)

	for _, sc := range scheduleCounts {
		stats.Schedules.Total += sc.Count
		if sc.IsActive {
			stats.Schedules.Active = sc.Count
		}
	}

	return stats, nil
}

func (s *DashboardService) getSummaryStats(ctx context.Context, workspaceID uuid.UUID, startOfDay, startOfWeek, startOfMonth time.Time) (*SummaryStats, error) {
	stats := &SummaryStats{}

	// Workflow counts by status
	var workflowCounts []struct {
		Status string `gorm:"column:status"`
		Count  int    `gorm:"column:count"`
	}
	s.db.WithContext(ctx).
		Model(&models.Workflow{}).
		Select("status, COUNT(*) as count").
		Where("workspace_id = ?", workspaceID).
		Group("status").
		Find(&workflowCounts)

	for _, wc := range workflowCounts {
		stats.TotalWorkflows += wc.Count
		switch wc.Status {
		case "active":
			stats.ActiveWorkflows = wc.Count
		case "inactive":
			stats.InactiveWorkflows = wc.Count
		case "draft":
			stats.DraftWorkflows = wc.Count
		}
	}

	// Execution counts
	s.db.WithContext(ctx).
		Model(&models.Execution{}).
		Where("workspace_id = ? AND created_at >= ?", workspaceID, startOfDay).
		Count(&stats.TotalExecutionsToday)

	s.db.WithContext(ctx).
		Model(&models.Execution{}).
		Where("workspace_id = ? AND created_at >= ?", workspaceID, startOfWeek).
		Count(&stats.TotalExecutionsWeek)

	s.db.WithContext(ctx).
		Model(&models.Execution{}).
		Where("workspace_id = ? AND created_at >= ?", workspaceID, startOfMonth).
		Count(&stats.TotalExecutionsMonth)

	// Running and queued
	s.db.WithContext(ctx).
		Model(&models.Execution{}).
		Where("workspace_id = ? AND status = ?", workspaceID, "running").
		Count(&stats.RunningExecutions)

	s.db.WithContext(ctx).
		Model(&models.Execution{}).
		Where("workspace_id = ? AND status = ?", workspaceID, "queued").
		Count(&stats.QueuedExecutions)

	// Success rate (last 30 days)
	var successRate struct {
		Total   int64 `gorm:"column:total"`
		Success int64 `gorm:"column:success"`
	}
	s.db.WithContext(ctx).
		Model(&models.Execution{}).
		Select("COUNT(*) as total, SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) as success").
		Where("workspace_id = ? AND created_at >= ?", workspaceID, startOfMonth).
		Find(&successRate)

	if successRate.Total > 0 {
		stats.SuccessRate = float64(successRate.Success) / float64(successRate.Total) * 100
	}

	// Average duration (completed executions)
	var avgDuration struct {
		AvgMs float64 `gorm:"column:avg_ms"`
	}
	s.db.WithContext(ctx).
		Model(&models.Execution{}).
		Select("AVG(EXTRACT(EPOCH FROM (completed_at - started_at)) * 1000) as avg_ms").
		Where("workspace_id = ? AND status = 'completed' AND started_at IS NOT NULL AND completed_at IS NOT NULL AND created_at >= ?", workspaceID, startOfMonth).
		Find(&avgDuration)
	stats.AvgDurationMs = int64(avgDuration.AvgMs)

	// Credentials count
	s.db.WithContext(ctx).
		Model(&models.Credential{}).
		Where("workspace_id = ?", workspaceID).
		Count(&stats.TotalCredentials)

	// Schedules count
	s.db.WithContext(ctx).
		Model(&models.Schedule{}).
		Where("workspace_id = ?", workspaceID).
		Count(&stats.TotalSchedules)

	s.db.WithContext(ctx).
		Model(&models.Schedule{}).
		Where("workspace_id = ? AND is_active = ?", workspaceID, true).
		Count(&stats.ActiveSchedules)

	return stats, nil
}

func (s *DashboardService) getRecentExecutions(ctx context.Context, workspaceID uuid.UUID, limit int) ([]ExecutionSummary, error) {
	var executions []models.Execution
	err := s.db.WithContext(ctx).
		Preload("Workflow").
		Where("workspace_id = ?", workspaceID).
		Order("created_at DESC").
		Limit(limit).
		Find(&executions).Error
	if err != nil {
		return nil, err
	}

	results := make([]ExecutionSummary, len(executions))
	for i, e := range executions {
		results[i] = ExecutionSummary{
			ID:           e.ID.String(),
			WorkflowID:   e.WorkflowID.String(),
			WorkflowName: e.Workflow.Name,
			Status:       e.Status,
			TriggerType:  e.TriggerType,
			CreatedAt:    e.CreatedAt.Unix(),
		}
		if e.StartedAt != nil && e.CompletedAt != nil {
			durationMs := e.CompletedAt.Sub(*e.StartedAt).Milliseconds()
			results[i].DurationMs = &durationMs
		}
		if e.StartedAt != nil {
			ts := e.StartedAt.Unix()
			results[i].StartedAt = &ts
		}
		if e.CompletedAt != nil {
			ts := e.CompletedAt.Unix()
			results[i].CompletedAt = &ts
		}
	}
	return results, nil
}

func (s *DashboardService) getTopWorkflows(ctx context.Context, workspaceID uuid.UUID, since time.Time, limit int) ([]WorkflowStats, error) {
	var results []struct {
		WorkflowID    uuid.UUID `gorm:"column:workflow_id"`
		Name          string    `gorm:"column:name"`
		Status        string    `gorm:"column:status"`
		ExecCount     int64     `gorm:"column:exec_count"`
		SuccessCount  int64     `gorm:"column:success_count"`
		FailedCount   int64     `gorm:"column:failed_count"`
		AvgDuration   float64   `gorm:"column:avg_duration"`
		LastExecuted  *time.Time `gorm:"column:last_executed"`
	}

	err := s.db.WithContext(ctx).
		Table("executions e").
		Select(`
			e.workflow_id,
			w.name,
			w.status,
			COUNT(*) as exec_count,
			SUM(CASE WHEN e.status = 'completed' THEN 1 ELSE 0 END) as success_count,
			SUM(CASE WHEN e.status = 'failed' THEN 1 ELSE 0 END) as failed_count,
			AVG(EXTRACT(EPOCH FROM (e.completed_at - e.started_at)) * 1000) as avg_duration,
			MAX(e.created_at) as last_executed
		`).
		Joins("JOIN workflows w ON w.id = e.workflow_id").
		Where("e.workspace_id = ? AND e.created_at >= ?", workspaceID, since).
		Group("e.workflow_id, w.name, w.status").
		Order("exec_count DESC").
		Limit(limit).
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	stats := make([]WorkflowStats, len(results))
	for i, r := range results {
		stats[i] = WorkflowStats{
			ID:             r.WorkflowID.String(),
			Name:           r.Name,
			Status:         r.Status,
			ExecutionCount: r.ExecCount,
			SuccessCount:   r.SuccessCount,
			FailedCount:    r.FailedCount,
			AvgDurationMs:  int64(r.AvgDuration),
		}
		if r.ExecCount > 0 {
			stats[i].SuccessRate = float64(r.SuccessCount) / float64(r.ExecCount) * 100
		}
		if r.LastExecuted != nil {
			ts := r.LastExecuted.Unix()
			stats[i].LastExecutedAt = &ts
		}
	}
	return stats, nil
}

func (s *DashboardService) getRecentFailures(ctx context.Context, workspaceID uuid.UUID, limit int) ([]FailureSummary, error) {
	var executions []models.Execution
	err := s.db.WithContext(ctx).
		Preload("Workflow").
		Where("workspace_id = ? AND status = ?", workspaceID, "failed").
		Order("created_at DESC").
		Limit(limit).
		Find(&executions).Error
	if err != nil {
		return nil, err
	}

	results := make([]FailureSummary, len(executions))
	for i, e := range executions {
		results[i] = FailureSummary{
			ID:           e.ID.String(),
			WorkflowID:   e.WorkflowID.String(),
			WorkflowName: e.Workflow.Name,
			FailedAt:     e.CreatedAt.Unix(),
		}
		if e.ErrorMessage != nil {
			results[i].ErrorMessage = *e.ErrorMessage
		}
		if e.ErrorNodeID != nil {
			results[i].ErrorNodeID = *e.ErrorNodeID
		}
	}
	return results, nil
}

func (s *DashboardService) getExecutionsByDay(ctx context.Context, workspaceID uuid.UUID, since time.Time) ([]DailyExecutions, error) {
	var results []struct {
		Date    time.Time `gorm:"column:date"`
		Total   int64     `gorm:"column:total"`
		Success int64     `gorm:"column:success"`
		Failed  int64     `gorm:"column:failed"`
	}

	err := s.db.WithContext(ctx).
		Model(&models.Execution{}).
		Select(`
			DATE(created_at) as date,
			COUNT(*) as total,
			SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) as success,
			SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failed
		`).
		Where("workspace_id = ? AND created_at >= ?", workspaceID, since).
		Group("DATE(created_at)").
		Order("date").
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	daily := make([]DailyExecutions, len(results))
	for i, r := range results {
		daily[i] = DailyExecutions{
			Date:    r.Date.Format("2006-01-02"),
			Total:   r.Total,
			Success: r.Success,
			Failed:  r.Failed,
		}
	}
	return daily, nil
}

func (s *DashboardService) getExecutionsByHour(ctx context.Context, workspaceID uuid.UUID) ([]HourlyExecutions, error) {
	since := time.Now().Add(-24 * time.Hour)

	var results []struct {
		Hour  int   `gorm:"column:hour"`
		Count int64 `gorm:"column:count"`
	}

	err := s.db.WithContext(ctx).
		Model(&models.Execution{}).
		Select("EXTRACT(HOUR FROM created_at)::int as hour, COUNT(*) as count").
		Where("workspace_id = ? AND created_at >= ?", workspaceID, since).
		Group("hour").
		Order("hour").
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	// Fill in missing hours with 0
	hourMap := make(map[int]int64)
	for _, r := range results {
		hourMap[r.Hour] = r.Count
	}

	hourly := make([]HourlyExecutions, 24)
	for h := 0; h < 24; h++ {
		hourly[h] = HourlyExecutions{
			Hour:  h,
			Count: hourMap[h],
		}
	}
	return hourly, nil
}

func (s *DashboardService) getUpcomingSchedules(ctx context.Context, workspaceID uuid.UUID, limit int) ([]ScheduleSummary, error) {
	var schedules []models.Schedule
	err := s.db.WithContext(ctx).
		Preload("Workflow").
		Where("workspace_id = ? AND is_active = ?", workspaceID, true).
		Order("next_run_at").
		Limit(limit).
		Find(&schedules).Error
	if err != nil {
		return nil, err
	}

	results := make([]ScheduleSummary, len(schedules))
	for i, s := range schedules {
		results[i] = ScheduleSummary{
			ID:             s.ID.String(),
			WorkflowID:     s.WorkflowID.String(),
			WorkflowName:   s.Workflow.Name,
			CronExpression: s.CronExpression,
			Timezone:       s.Timezone,
			IsActive:       s.IsActive,
		}
		if s.NextRunAt != nil {
			ts := s.NextRunAt.Unix()
			results[i].NextRunAt = &ts
		}
	}
	return results, nil
}

func (s *DashboardService) getExecutionsByStatus(ctx context.Context, workspaceID uuid.UUID, since time.Time) ([]StatusCount, error) {
	var results []struct {
		Status string `gorm:"column:status"`
		Count  int64  `gorm:"column:count"`
	}

	err := s.db.WithContext(ctx).
		Model(&models.Execution{}).
		Select("status, COUNT(*) as count").
		Where("workspace_id = ? AND created_at >= ?", workspaceID, since).
		Group("status").
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	counts := make([]StatusCount, len(results))
	for i, r := range results {
		counts[i] = StatusCount{
			Status: r.Status,
			Count:  r.Count,
		}
	}
	return counts, nil
}

func (s *DashboardService) getTriggerTypeStats(ctx context.Context, workspaceID uuid.UUID, since time.Time) ([]TriggerTypeCount, error) {
	var results []struct {
		TriggerType string `gorm:"column:trigger_type"`
		Count       int64  `gorm:"column:count"`
	}

	err := s.db.WithContext(ctx).
		Model(&models.Execution{}).
		Select("trigger_type, COUNT(*) as count").
		Where("workspace_id = ? AND created_at >= ?", workspaceID, since).
		Group("trigger_type").
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	counts := make([]TriggerTypeCount, len(results))
	for i, r := range results {
		counts[i] = TriggerTypeCount{
			TriggerType: r.TriggerType,
			Count:       r.Count,
		}
	}
	return counts, nil
}
