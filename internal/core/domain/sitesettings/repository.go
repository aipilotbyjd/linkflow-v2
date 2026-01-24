package sitesettings

import (
	"context"
)

// Repository defines the interface for site settings persistence
type Repository interface {
	// Get returns the site settings (creates default if not exists)
	Get(ctx context.Context) (*SiteSettings, error)

	// Update updates the site settings
	Update(ctx context.Context, settings *SiteSettings) error

	// GetDisabledNodes returns just the disabled nodes list
	GetDisabledNodes(ctx context.Context) ([]string, error)

	// SetDisabledNodes updates only the disabled nodes list
	SetDisabledNodes(ctx context.Context, nodes []string) error
}
