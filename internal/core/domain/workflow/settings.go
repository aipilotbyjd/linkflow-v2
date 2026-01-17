package workflow

import "encoding/json"

// Settings represents workflow settings
type Settings struct {
	TimeoutSeconds    int               `json:"timeout_seconds,omitempty"`
	RetryOnFailure    bool              `json:"retry_on_failure,omitempty"`
	MaxRetries        int               `json:"max_retries,omitempty"`
	RetryDelaySeconds int               `json:"retry_delay_seconds,omitempty"`
	SaveExecutionData bool              `json:"save_execution_data,omitempty"`
	Timezone          string            `json:"timezone,omitempty"`
	ErrorWorkflowID   string            `json:"error_workflow_id,omitempty"`
	Variables         map[string]string `json:"variables,omitempty"`
	Tags              []string          `json:"tags,omitempty"`
	ExecutionOrder    string            `json:"execution_order,omitempty"` // sequential, parallel
	ConcurrencyLimit  int               `json:"concurrency_limit,omitempty"`
	DeduplicationKey  string            `json:"deduplication_key,omitempty"`
	DeduplicationTTL  int               `json:"deduplication_ttl,omitempty"` // seconds
}

// DefaultSettings returns default workflow settings
func DefaultSettings() Settings {
	return Settings{
		TimeoutSeconds:    3600, // 1 hour
		RetryOnFailure:    false,
		MaxRetries:        0,
		RetryDelaySeconds: 60,
		SaveExecutionData: true,
		Timezone:          "UTC",
		ExecutionOrder:    "sequential",
		ConcurrencyLimit:  1,
	}
}

// Merge merges another settings into this one (other takes precedence for non-zero values)
func (s Settings) Merge(other Settings) Settings {
	if other.TimeoutSeconds > 0 {
		s.TimeoutSeconds = other.TimeoutSeconds
	}
	if other.RetryOnFailure {
		s.RetryOnFailure = other.RetryOnFailure
	}
	if other.MaxRetries > 0 {
		s.MaxRetries = other.MaxRetries
	}
	if other.RetryDelaySeconds > 0 {
		s.RetryDelaySeconds = other.RetryDelaySeconds
	}
	if other.Timezone != "" {
		s.Timezone = other.Timezone
	}
	if other.ErrorWorkflowID != "" {
		s.ErrorWorkflowID = other.ErrorWorkflowID
	}
	if other.Variables != nil {
		if s.Variables == nil {
			s.Variables = make(map[string]string)
		}
		for k, v := range other.Variables {
			s.Variables[k] = v
		}
	}
	if other.Tags != nil {
		s.Tags = other.Tags
	}
	if other.ExecutionOrder != "" {
		s.ExecutionOrder = other.ExecutionOrder
	}
	if other.ConcurrencyLimit > 0 {
		s.ConcurrencyLimit = other.ConcurrencyLimit
	}
	if other.DeduplicationKey != "" {
		s.DeduplicationKey = other.DeduplicationKey
	}
	if other.DeduplicationTTL > 0 {
		s.DeduplicationTTL = other.DeduplicationTTL
	}
	return s
}

// SettingsFromJSON parses settings from JSON
func SettingsFromJSON(data []byte) (Settings, error) {
	var settings Settings
	if err := json.Unmarshal(data, &settings); err != nil {
		return DefaultSettings(), err
	}
	return settings, nil
}

// ToJSON converts settings to JSON
func (s Settings) ToJSON() ([]byte, error) {
	return json.Marshal(s)
}

// GetVariable returns a variable value
func (s Settings) GetVariable(key string) (string, bool) {
	if s.Variables == nil {
		return "", false
	}
	val, ok := s.Variables[key]
	return val, ok
}

// SetVariable sets a variable value
func (s *Settings) SetVariable(key, value string) {
	if s.Variables == nil {
		s.Variables = make(map[string]string)
	}
	s.Variables[key] = value
}

// HasTag checks if a tag exists
func (s Settings) HasTag(tag string) bool {
	for _, t := range s.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// AddTag adds a tag
func (s *Settings) AddTag(tag string) {
	if !s.HasTag(tag) {
		s.Tags = append(s.Tags, tag)
	}
}

// RemoveTag removes a tag
func (s *Settings) RemoveTag(tag string) {
	for i, t := range s.Tags {
		if t == tag {
			s.Tags = append(s.Tags[:i], s.Tags[i+1:]...)
			return
		}
	}
}
