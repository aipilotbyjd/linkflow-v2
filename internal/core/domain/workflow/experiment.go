package workflow

import (
	"time"

	"github.com/google/uuid"
)

// Experiment represents an A/B test experiment for workflows
type Experiment struct {
	ID            uuid.UUID           `json:"id" gorm:"type:uuid;primaryKey"`
	WorkspaceID   uuid.UUID           `json:"workspace_id" gorm:"type:uuid;index;not null"`
	WorkflowID    uuid.UUID           `json:"workflow_id" gorm:"type:uuid;index;not null"`
	Name          string              `json:"name" gorm:"size:255;not null"`
	Description   *string             `json:"description,omitempty"`
	Status        ExperimentStatus    `json:"status" gorm:"size:20;not null;default:draft"`
	Variants      []ExperimentVariant `json:"variants" gorm:"-"`
	TrafficSplit  TrafficSplit        `json:"traffic_split" gorm:"type:jsonb"`
	SuccessMetric string              `json:"success_metric" gorm:"size:50"` // completion_rate, duration, error_rate
	StartedAt     *time.Time          `json:"started_at,omitempty"`
	EndedAt       *time.Time          `json:"ended_at,omitempty"`
	CreatedBy     uuid.UUID           `json:"created_by" gorm:"type:uuid;not null"`
	CreatedAt     time.Time           `json:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at"`
}

// ExperimentStatus represents the status of an experiment
type ExperimentStatus string

const (
	ExperimentStatusDraft     ExperimentStatus = "draft"
	ExperimentStatusRunning   ExperimentStatus = "running"
	ExperimentStatusPaused    ExperimentStatus = "paused"
	ExperimentStatusCompleted ExperimentStatus = "completed"
	ExperimentStatusCancelled ExperimentStatus = "canceled"
)

// ExperimentVariant represents a variant in an A/B test
type ExperimentVariant struct {
	ID           uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	ExperimentID uuid.UUID      `json:"experiment_id" gorm:"type:uuid;index;not null"`
	Name         string         `json:"name" gorm:"size:50;not null"` // control, variant_a, variant_b
	Description  *string        `json:"description,omitempty"`
	WorkflowID   uuid.UUID      `json:"workflow_id" gorm:"type:uuid;not null"` // Can be same or different workflow
	Version      int            `json:"version"`                               // Specific version or 0 for latest
	Weight       int            `json:"weight" gorm:"default:50"`              // Traffic weight percentage
	IsControl    bool           `json:"is_control" gorm:"default:false"`
	Metrics      VariantMetrics `json:"metrics" gorm:"type:jsonb"`
}

// TrafficSplit defines how traffic is split between variants
type TrafficSplit struct {
	Type      string        `json:"type"` // percentage, user_segment, time_based
	Rules     []TrafficRule `json:"rules,omitempty"`
	Sticky    bool          `json:"sticky"`     // Same user always gets same variant
	StickyKey string        `json:"sticky_key"` // Field to use for sticky assignment (user_id, session_id)
}

// TrafficRule defines a traffic routing rule
type TrafficRule struct {
	VariantID uuid.UUID              `json:"variant_id"`
	Condition map[string]interface{} `json:"condition,omitempty"` // e.g., {"user_segment": "premium"}
	Weight    int                    `json:"weight"`
}

// VariantMetrics holds metrics for a variant
type VariantMetrics struct {
	TotalExecutions      int64   `json:"total_executions"`
	SuccessfulExecutions int64   `json:"successful_executions"`
	FailedExecutions     int64   `json:"failed_executions"`
	AverageDurationMs    float64 `json:"average_duration_ms"`
	P50DurationMs        float64 `json:"p50_duration_ms"`
	P95DurationMs        float64 `json:"p95_duration_ms"`
	P99DurationMs        float64 `json:"p99_duration_ms"`
	SuccessRate          float64 `json:"success_rate"`
	ErrorRate            float64 `json:"error_rate"`
	ConfidenceLevel      float64 `json:"confidence_level"` // Statistical confidence
	IsWinner             bool    `json:"is_winner"`
}

// NewExperiment creates a new experiment
func NewExperiment(workspaceID, workflowID, createdBy uuid.UUID, name string) *Experiment {
	return &Experiment{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		WorkflowID:  workflowID,
		Name:        name,
		Status:      ExperimentStatusDraft,
		TrafficSplit: TrafficSplit{
			Type:      "percentage",
			Sticky:    true,
			StickyKey: "user_id",
		},
		SuccessMetric: "completion_rate",
		CreatedBy:     createdBy,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// AddVariant adds a variant to the experiment
func (e *Experiment) AddVariant(name string, workflowID uuid.UUID, version int, weight int, isControl bool) *ExperimentVariant {
	variant := &ExperimentVariant{
		ID:           uuid.New(),
		ExperimentID: e.ID,
		Name:         name,
		WorkflowID:   workflowID,
		Version:      version,
		Weight:       weight,
		IsControl:    isControl,
	}
	e.Variants = append(e.Variants, *variant)
	return variant
}

// Start starts the experiment
func (e *Experiment) Start() error {
	if len(e.Variants) < 2 {
		return ErrInsufficientVariants
	}
	e.Status = ExperimentStatusRunning
	now := time.Now()
	e.StartedAt = &now
	e.UpdatedAt = now
	return nil
}

// Pause pauses the experiment
func (e *Experiment) Pause() {
	e.Status = ExperimentStatusPaused
	e.UpdatedAt = time.Now()
}

// Resume resumes a paused experiment
func (e *Experiment) Resume() {
	e.Status = ExperimentStatusRunning
	e.UpdatedAt = time.Now()
}

// Complete completes the experiment
func (e *Experiment) Complete() {
	e.Status = ExperimentStatusCompleted
	now := time.Now()
	e.EndedAt = &now
	e.UpdatedAt = now
}

// Cancel cancels the experiment
func (e *Experiment) Cancel() {
	e.Status = ExperimentStatusCancelled
	now := time.Now()
	e.EndedAt = &now
	e.UpdatedAt = now
}

// SelectVariant selects a variant based on traffic split rules
func (e *Experiment) SelectVariant(stickyValue string) *ExperimentVariant {
	if len(e.Variants) == 0 {
		return nil
	}

	// If sticky, use consistent hashing
	if e.TrafficSplit.Sticky && stickyValue != "" {
		hash := hashString(stickyValue + e.ID.String())
		totalWeight := 0
		for _, v := range e.Variants {
			totalWeight += v.Weight
		}

		target := hash % totalWeight
		cumWeight := 0
		for i := range e.Variants {
			cumWeight += e.Variants[i].Weight
			if target < cumWeight {
				return &e.Variants[i]
			}
		}
	}

	// Random selection based on weight
	// Implementation would use random selection with weights
	return &e.Variants[0]
}

// GetWinner determines the winning variant based on metrics
func (e *Experiment) GetWinner() *ExperimentVariant {
	var winner *ExperimentVariant
	bestMetric := 0.0

	for i := range e.Variants {
		v := &e.Variants[i]
		var metric float64

		switch e.SuccessMetric {
		case "completion_rate":
			metric = v.Metrics.SuccessRate
		case "duration":
			if v.Metrics.AverageDurationMs > 0 {
				metric = 1.0 / v.Metrics.AverageDurationMs
			}
		case "error_rate":
			metric = 1.0 - v.Metrics.ErrorRate
		}

		if metric > bestMetric && v.Metrics.ConfidenceLevel >= 0.95 {
			bestMetric = metric
			winner = v
		}
	}

	return winner
}

// Simple hash function for consistent variant selection
func hashString(s string) int {
	h := 0
	for _, c := range s {
		h = 31*h + int(c)
	}
	if h < 0 {
		h = -h
	}
	return h
}
