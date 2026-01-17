package dashboard

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/execution"
	"github.com/linkflow-ai/linkflow/internal/core/domain/schedule"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workflow"
)

type DashboardHandler struct {
	workflowRepo  workflow.Repository
	executionRepo execution.Repository
	scheduleRepo  schedule.Repository
}

func NewDashboardHandler(
	workflowRepo workflow.Repository,
	executionRepo execution.Repository,
	scheduleRepo schedule.Repository,
) *DashboardHandler {
	return &DashboardHandler{
		workflowRepo:  workflowRepo,
		executionRepo: executionRepo,
		scheduleRepo:  scheduleRepo,
	}
}

type DashboardResponse struct {
	Summary            SummaryStats       `json:"summary"`
	RecentExecutions   []ExecutionSummary `json:"recent_executions"`
	TopWorkflows       []WorkflowStats    `json:"top_workflows"`
	RecentFailures     []ExecutionSummary `json:"recent_failures"`
	ExecutionsByDay    []DailyStats       `json:"executions_by_day"`
	ExecutionsByHour   []HourlyStats      `json:"executions_by_hour"`
	UpcomingSchedules  []ScheduleSummary  `json:"upcoming_schedules"`
	StatusDistribution map[string]int64   `json:"status_distribution"`
	TriggerTypeStats   map[string]int64   `json:"trigger_type_stats"`
}

type SummaryStats struct {
	TotalWorkflows   int64   `json:"total_workflows"`
	ActiveWorkflows  int64   `json:"active_workflows"`
	TotalExecutions  int64   `json:"total_executions"`
	ExecutionsToday  int64   `json:"executions_today"`
	ExecutionsWeek   int64   `json:"executions_week"`
	ExecutionsMonth  int64   `json:"executions_month"`
	SuccessRate      float64 `json:"success_rate"`
	AvgDurationMs    int64   `json:"avg_duration_ms"`
	TotalCredentials int64   `json:"total_credentials"`
	TotalSchedules   int64   `json:"total_schedules"`
	ActiveSchedules  int64   `json:"active_schedules"`
}

type ExecutionSummary struct {
	ID           uuid.UUID  `json:"id"`
	WorkflowID   uuid.UUID  `json:"workflow_id"`
	WorkflowName string     `json:"workflow_name"`
	Status       string     `json:"status"`
	TriggerType  string     `json:"trigger_type"`
	StartedAt    time.Time  `json:"started_at"`
	FinishedAt   *time.Time `json:"finished_at,omitempty"`
	DurationMs   int64      `json:"duration_ms"`
	Error        string     `json:"error,omitempty"`
}

type WorkflowStats struct {
	ID             uuid.UUID  `json:"id"`
	Name           string     `json:"name"`
	ExecutionCount int64      `json:"execution_count"`
	SuccessCount   int64      `json:"success_count"`
	FailureCount   int64      `json:"failure_count"`
	SuccessRate    float64    `json:"success_rate"`
	AvgDurationMs  int64      `json:"avg_duration_ms"`
	LastExecutedAt *time.Time `json:"last_executed_at,omitempty"`
}

type DailyStats struct {
	Date      string `json:"date"`
	Total     int64  `json:"total"`
	Completed int64  `json:"completed"`
	Failed    int64  `json:"failed"`
	Canceled  int64  `json:"canceled"`
}

type HourlyStats struct {
	Hour  int   `json:"hour"`
	Count int64 `json:"count"`
}

type ScheduleSummary struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	WorkflowID   uuid.UUID `json:"workflow_id"`
	WorkflowName string    `json:"workflow_name"`
	NextRunAt    time.Time `json:"next_run_at"`
	Cron         string    `json:"cron"`
}

func (h *DashboardHandler) Handle(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		common.BadRequest(w, "workspace context required")
		return
	}

	period := r.URL.Query().Get("period")
	if period == "" {
		period = "7d"
	}

	var days int
	switch period {
	case "7d":
		days = 7
	case "30d":
		days = 30
	case "90d":
		days = 90
	default:
		days = 7
	}

	now := time.Now()
	startDate := now.AddDate(0, 0, -days)
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekStart := now.AddDate(0, 0, -7)
	monthStart := now.AddDate(0, -1, 0)

	// Build response with available data
	response := DashboardResponse{
		Summary: SummaryStats{
			TotalWorkflows:   0,
			ActiveWorkflows:  0,
			TotalExecutions:  0,
			ExecutionsToday:  0,
			ExecutionsWeek:   0,
			ExecutionsMonth:  0,
			SuccessRate:      0,
			AvgDurationMs:    0,
			TotalCredentials: 0,
			TotalSchedules:   0,
			ActiveSchedules:  0,
		},
		RecentExecutions:   []ExecutionSummary{},
		TopWorkflows:       []WorkflowStats{},
		RecentFailures:     []ExecutionSummary{},
		ExecutionsByDay:    make([]DailyStats, 0),
		ExecutionsByHour:   make([]HourlyStats, 24),
		UpcomingSchedules:  []ScheduleSummary{},
		StatusDistribution: make(map[string]int64),
		TriggerTypeStats:   make(map[string]int64),
	}

	// Get workflow stats
	workflows, totalWf, _ := h.workflowRepo.FindByWorkspaceID(r.Context(), wsCtx.WorkspaceID, nil)
	response.Summary.TotalWorkflows = totalWf

	activeCount := int64(0)
	for _, wf := range workflows {
		if wf.Status == workflow.StatusActive {
			activeCount++
		}
	}
	response.Summary.ActiveWorkflows = activeCount

	// Get execution stats
	executions, totalExec, _ := h.executionRepo.FindByWorkspaceID(r.Context(), wsCtx.WorkspaceID, nil)
	response.Summary.TotalExecutions = totalExec

	// Count executions by time period and build stats
	var successCount, failCount int64
	var totalDuration int64
	for _, exec := range executions {
		if exec.StartedAt != nil && exec.StartedAt.After(todayStart) {
			response.Summary.ExecutionsToday++
		}
		if exec.StartedAt != nil && exec.StartedAt.After(weekStart) {
			response.Summary.ExecutionsWeek++
		}
		if exec.StartedAt != nil && exec.StartedAt.After(monthStart) {
			response.Summary.ExecutionsMonth++
		}
		if exec.Status == execution.StatusCompleted {
			successCount++
		}
		if exec.Status == execution.StatusFailed {
			failCount++
		}
		if exec.CompletedAt != nil && exec.StartedAt != nil {
			totalDuration += exec.CompletedAt.Sub(*exec.StartedAt).Milliseconds()
		}

		// Status distribution
		response.StatusDistribution[string(exec.Status)]++

		// Trigger type stats
		response.TriggerTypeStats[exec.TriggerType]++
	}

	if totalExec > 0 {
		response.Summary.SuccessRate = float64(successCount) / float64(totalExec) * 100
		response.Summary.AvgDurationMs = totalDuration / totalExec
	}

	// Build recent executions (last 10)
	count := 10
	if len(executions) < count {
		count = len(executions)
	}
	for i := 0; i < count; i++ {
		exec := executions[i]
		var durationMs int64
		if exec.CompletedAt != nil && exec.StartedAt != nil {
			durationMs = exec.CompletedAt.Sub(*exec.StartedAt).Milliseconds()
		}

		wfName := ""
		for _, wf := range workflows {
			if wf.ID == exec.WorkflowID {
				wfName = wf.Name
				break
			}
		}

		var startedAt time.Time
		if exec.StartedAt != nil {
			startedAt = *exec.StartedAt
		}
		var errMsg string
		if exec.ErrorMessage != nil {
			errMsg = *exec.ErrorMessage
		}

		response.RecentExecutions = append(response.RecentExecutions, ExecutionSummary{
			ID:           exec.ID,
			WorkflowID:   exec.WorkflowID,
			WorkflowName: wfName,
			Status:       string(exec.Status),
			TriggerType:  exec.TriggerType,
			StartedAt:    startedAt,
			FinishedAt:   exec.CompletedAt,
			DurationMs:   durationMs,
			Error:        errMsg,
		})
	}

	// Build recent failures
	for _, exec := range executions {
		if exec.Status == execution.StatusFailed && len(response.RecentFailures) < 10 {
			var durationMs int64
			if exec.CompletedAt != nil && exec.StartedAt != nil {
				durationMs = exec.CompletedAt.Sub(*exec.StartedAt).Milliseconds()
			}

			wfName := ""
			for _, wf := range workflows {
				if wf.ID == exec.WorkflowID {
					wfName = wf.Name
					break
				}
			}

			var startedAt time.Time
			if exec.StartedAt != nil {
				startedAt = *exec.StartedAt
			}
			var errMsg string
			if exec.ErrorMessage != nil {
				errMsg = *exec.ErrorMessage
			}

			response.RecentFailures = append(response.RecentFailures, ExecutionSummary{
				ID:           exec.ID,
				WorkflowID:   exec.WorkflowID,
				WorkflowName: wfName,
				Status:       string(exec.Status),
				TriggerType:  exec.TriggerType,
				StartedAt:    startedAt,
				FinishedAt:   exec.CompletedAt,
				DurationMs:   durationMs,
				Error:        errMsg,
			})
		}
	}

	// Build executions by day
	dailyMap := make(map[string]*DailyStats)
	for i := 0; i < days; i++ {
		date := startDate.AddDate(0, 0, i).Format("2006-01-02")
		dailyMap[date] = &DailyStats{Date: date}
	}
	for _, exec := range executions {
		if exec.StartedAt != nil && exec.StartedAt.After(startDate) {
			date := exec.StartedAt.Format("2006-01-02")
			if stats, ok := dailyMap[date]; ok {
				stats.Total++
				switch exec.Status {
				case execution.StatusCompleted:
					stats.Completed++
				case execution.StatusFailed:
					stats.Failed++
				case execution.StatusCancelled:
					stats.Canceled++
				}
			}
		}
	}
	for i := 0; i < days; i++ {
		date := startDate.AddDate(0, 0, i).Format("2006-01-02")
		if stats, ok := dailyMap[date]; ok {
			response.ExecutionsByDay = append(response.ExecutionsByDay, *stats)
		}
	}

	// Build hourly distribution
	for i := 0; i < 24; i++ {
		response.ExecutionsByHour[i] = HourlyStats{Hour: i, Count: 0}
	}
	for _, exec := range executions {
		if exec.StartedAt != nil && exec.StartedAt.After(startDate) {
			hour := exec.StartedAt.Hour()
			response.ExecutionsByHour[hour].Count++
		}
	}

	// Get schedule stats
	schedules, totalSch, _ := h.scheduleRepo.FindByWorkspaceID(r.Context(), wsCtx.WorkspaceID, nil)
	response.Summary.TotalSchedules = totalSch

	activeSchCount := int64(0)
	for _, sch := range schedules {
		if sch.IsActive {
			activeSchCount++
			if sch.NextRunAt != nil && len(response.UpcomingSchedules) < 5 {
				wfName := ""
				for _, wf := range workflows {
					if wf.ID == sch.WorkflowID {
						wfName = wf.Name
						break
					}
				}
				response.UpcomingSchedules = append(response.UpcomingSchedules, ScheduleSummary{
					ID:           sch.ID,
					Name:         sch.Name,
					WorkflowID:   sch.WorkflowID,
					WorkflowName: wfName,
					NextRunAt:    *sch.NextRunAt,
					Cron:         sch.CronExpression,
				})
			}
		}
	}
	response.Summary.ActiveSchedules = activeSchCount

	common.Success(w, response)
}
