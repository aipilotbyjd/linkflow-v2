package sitesettings

import (
	"time"

	"github.com/lib/pq"
)

// SiteSettings represents global site-wide settings
// There should only be one row in this table (singleton pattern)
type SiteSettings struct {
	ID            uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	DisabledNodes pq.StringArray `gorm:"type:text[];default:'{}'" json:"disabled_nodes"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

func (SiteSettings) TableName() string {
	return "site_settings"
}

// NewSiteSettings creates a new site settings with defaults
func NewSiteSettings() *SiteSettings {
	return &SiteSettings{
		ID:            1, // Singleton - always ID 1
		DisabledNodes: []string{},
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// DisableNode adds a node type to the disabled list
func (s *SiteSettings) DisableNode(nodeType string) {
	// Check if already disabled
	for _, n := range s.DisabledNodes {
		if n == nodeType {
			return
		}
	}
	s.DisabledNodes = append(s.DisabledNodes, nodeType)
	s.UpdatedAt = time.Now()
}

// EnableNode removes a node type from the disabled list
func (s *SiteSettings) EnableNode(nodeType string) {
	var updated []string
	for _, n := range s.DisabledNodes {
		if n != nodeType {
			updated = append(updated, n)
		}
	}
	s.DisabledNodes = updated
	s.UpdatedAt = time.Now()
}

// IsNodeDisabled checks if a node type is disabled
func (s *SiteSettings) IsNodeDisabled(nodeType string) bool {
	for _, n := range s.DisabledNodes {
		if n == nodeType {
			return true
		}
	}
	return false
}

// SetDisabledNodes sets the full list of disabled nodes
func (s *SiteSettings) SetDisabledNodes(nodes []string) {
	s.DisabledNodes = nodes
	s.UpdatedAt = time.Now()
}

// GetDisabledNodes returns the list of disabled node types
func (s *SiteSettings) GetDisabledNodes() []string {
	if s.DisabledNodes == nil {
		return []string{}
	}
	return s.DisabledNodes
}
