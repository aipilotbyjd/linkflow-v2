package billing

import (
	"sync"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/billing"
)

// AlertService handles usage alerts and notifications
type AlertService struct {
	mu              sync.RWMutex
	triggeredAlerts map[string]bool // Track which alerts already triggered this period
	alertQueue      []AlertEvent    // Queue of alerts to send
}

// AlertEvent represents an alert to be sent
type AlertEvent struct {
	WorkspaceID uuid.UUID
	AlertType   billing.UsageAlertType
	Threshold   int
	Current     int64
	Limit       int64
	Percentage  float64
	Message     string
	IsOverage   bool
}

func NewAlertService() *AlertService {
	return &AlertService{
		triggeredAlerts: make(map[string]bool),
		alertQueue:      make([]AlertEvent, 0),
	}
}

// CheckThresholds checks if any alert thresholds have been crossed
func (s *AlertService) CheckThresholds(workspaceID uuid.UUID, alertType billing.UsageAlertType, percentage float64, current, limit int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	thresholds := []int{50, 75, 90, 100}
	
	for _, threshold := range thresholds {
		if percentage >= float64(threshold) {
			key := s.alertKey(workspaceID, alertType, threshold)
			if !s.triggeredAlerts[key] {
				s.triggeredAlerts[key] = true
				s.queueAlert(workspaceID, alertType, threshold, current, limit, percentage, false)
			}
		}
	}
}

// TriggerAlert triggers a specific alert
func (s *AlertService) TriggerAlert(workspaceID uuid.UUID, alertType billing.UsageAlertType, threshold int, current, limit int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	percentage := float64(current) / float64(limit) * 100
	s.queueAlert(workspaceID, alertType, threshold, current, limit, percentage, false)
}

// TriggerOverageAlert triggers an overage alert
func (s *AlertService) TriggerOverageAlert(workspaceID uuid.UUID, alertType billing.UsageAlertType, overageAmount int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	key := s.alertKey(workspaceID, alertType, 999) // Special key for overage
	if !s.triggeredAlerts[key] {
		s.triggeredAlerts[key] = true
		s.alertQueue = append(s.alertQueue, AlertEvent{
			WorkspaceID: workspaceID,
			AlertType:   alertType,
			Threshold:   100,
			Current:     overageAmount,
			IsOverage:   true,
			Message:     "You are now in overage billing mode. Additional charges will apply.",
		})
	}
}

// GetPendingAlerts returns and clears pending alerts
func (s *AlertService) GetPendingAlerts() []AlertEvent {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	alerts := s.alertQueue
	s.alertQueue = make([]AlertEvent, 0)
	return alerts
}

// ResetPeriod resets alert tracking for new billing period
func (s *AlertService) ResetPeriod(workspaceID uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Clear all alerts for this workspace
	for key := range s.triggeredAlerts {
		// Keys start with workspace ID
		if len(key) > 36 && key[:36] == workspaceID.String() {
			delete(s.triggeredAlerts, key)
		}
	}
}

func (s *AlertService) alertKey(workspaceID uuid.UUID, alertType billing.UsageAlertType, threshold int) string {
	return workspaceID.String() + ":" + string(alertType) + ":" + string(rune(threshold))
}

func (s *AlertService) queueAlert(workspaceID uuid.UUID, alertType billing.UsageAlertType, threshold int, current, limit int64, percentage float64, isOverage bool) {
	message := s.generateMessage(alertType, threshold, percentage)
	
	s.alertQueue = append(s.alertQueue, AlertEvent{
		WorkspaceID: workspaceID,
		AlertType:   alertType,
		Threshold:   threshold,
		Current:     current,
		Limit:       limit,
		Percentage:  percentage,
		Message:     message,
		IsOverage:   isOverage,
	})
}

func (s *AlertService) generateMessage(alertType billing.UsageAlertType, threshold int, percentage float64) string {
	resource := string(alertType)
	
	switch threshold {
	case 50:
		return "You've used 50% of your monthly " + resource + ". Keep an eye on your usage."
	case 75:
		return "Warning: You've used 75% of your monthly " + resource + ". Consider upgrading soon."
	case 90:
		return "Critical: You've used 90% of your monthly " + resource + ". Upgrade now to avoid interruption."
	case 100:
		return "You've reached 100% of your monthly " + resource + " limit. Overage charges may apply."
	default:
		return "Usage alert for " + resource
	}
}
