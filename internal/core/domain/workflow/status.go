package workflow

// Status represents workflow status
type Status string

const (
	StatusDraft    Status = "draft"
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
	StatusArchived Status = "archived"
)

func (s Status) String() string {
	return string(s)
}

func (s Status) IsValid() bool {
	switch s {
	case StatusDraft, StatusActive, StatusInactive, StatusArchived:
		return true
	default:
		return false
	}
}

func ParseStatus(s string) (Status, bool) {
	status := Status(s)
	return status, status.IsValid()
}

// CanTransitionTo checks if workflow can transition to target status
func (s Status) CanTransitionTo(target Status) bool {
	switch s {
	case StatusDraft:
		return target == StatusActive || target == StatusArchived
	case StatusActive:
		return target == StatusInactive || target == StatusArchived
	case StatusInactive:
		return target == StatusActive || target == StatusArchived || target == StatusDraft
	case StatusArchived:
		return target == StatusDraft
	default:
		return false
	}
}

// TriggerType represents workflow trigger types
type TriggerType string

const (
	TriggerManual      TriggerType = "manual"
	TriggerSchedule    TriggerType = "schedule"
	TriggerWebhook     TriggerType = "webhook"
	TriggerAPI         TriggerType = "api"
	TriggerSubWorkflow TriggerType = "sub_workflow"
	TriggerReplay      TriggerType = "replay"
	TriggerError       TriggerType = "error_workflow"
)

func (t TriggerType) String() string {
	return string(t)
}

func (t TriggerType) IsValid() bool {
	switch t {
	case TriggerManual, TriggerSchedule, TriggerWebhook, TriggerAPI, TriggerSubWorkflow, TriggerReplay, TriggerError:
		return true
	default:
		return false
	}
}

// ErrorTrigger represents when to trigger error workflow
type ErrorTrigger string

const (
	ErrorTriggerOnFailure ErrorTrigger = "on_failure"
	ErrorTriggerOnTimeout ErrorTrigger = "on_timeout"
	ErrorTriggerOnAll     ErrorTrigger = "on_all"
)

func (e ErrorTrigger) String() string {
	return string(e)
}
