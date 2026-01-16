package user

// Status represents user account status
type Status string

const (
	StatusActive    Status = "active"
	StatusSuspended Status = "suspended"
	StatusDeleted   Status = "deleted"
	StatusPending   Status = "pending"
)

func (s Status) String() string {
	return string(s)
}

func (s Status) IsValid() bool {
	switch s {
	case StatusActive, StatusSuspended, StatusDeleted, StatusPending:
		return true
	default:
		return false
	}
}

func ParseStatus(s string) (Status, bool) {
	status := Status(s)
	return status, status.IsValid()
}
