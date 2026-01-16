package execution

// Status represents execution status
type Status string

const (
	StatusQueued    Status = "queued"
	StatusRunning   Status = "running"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
	StatusCancelled Status = "cancelled"
	StatusTimeout   Status = "timeout"
	StatusWaiting   Status = "waiting"
)

func (s Status) String() string {
	return string(s)
}

func (s Status) IsValid() bool {
	switch s {
	case StatusQueued, StatusRunning, StatusCompleted, StatusFailed, StatusCancelled, StatusTimeout, StatusWaiting:
		return true
	default:
		return false
	}
}

func ParseStatus(s string) (Status, bool) {
	status := Status(s)
	return status, status.IsValid()
}

// IsTerminal returns true if this is a terminal status
func (s Status) IsTerminal() bool {
	switch s {
	case StatusCompleted, StatusFailed, StatusCancelled, StatusTimeout:
		return true
	default:
		return false
	}
}

// IsSuccess returns true if execution was successful
func (s Status) IsSuccess() bool {
	return s == StatusCompleted
}

// IsFailure returns true if execution failed
func (s Status) IsFailure() bool {
	return s == StatusFailed || s == StatusTimeout
}

// NodeStatus represents node execution status
type NodeStatus string

const (
	NodeStatusPending   NodeStatus = "pending"
	NodeStatusRunning   NodeStatus = "running"
	NodeStatusCompleted NodeStatus = "completed"
	NodeStatusFailed    NodeStatus = "failed"
	NodeStatusSkipped   NodeStatus = "skipped"
)

func (s NodeStatus) String() string {
	return string(s)
}

func (s NodeStatus) IsValid() bool {
	switch s {
	case NodeStatusPending, NodeStatusRunning, NodeStatusCompleted, NodeStatusFailed, NodeStatusSkipped:
		return true
	default:
		return false
	}
}

// IsTerminal returns true if this is a terminal status
func (s NodeStatus) IsTerminal() bool {
	switch s {
	case NodeStatusCompleted, NodeStatusFailed, NodeStatusSkipped:
		return true
	default:
		return false
	}
}

// LogLevel represents log entry severity
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

func (l LogLevel) String() string {
	return string(l)
}
