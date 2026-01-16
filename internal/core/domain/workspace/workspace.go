package workspace

import (
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
	"gorm.io/gorm"
)

// Workspace entity (aggregate root)
type Workspace struct {
	ID               uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OwnerID          uuid.UUID      `gorm:"type:uuid;index;not null" json:"owner_id"`
	Name             string         `gorm:"size:100;not null" json:"name"`
	Slug             string         `gorm:"size:100;uniqueIndex;not null" json:"slug"`
	Description      *string        `gorm:"type:text" json:"description,omitempty"`
	LogoURL          *string        `gorm:"size:500" json:"logo_url,omitempty"`
	Website          *string        `gorm:"size:255" json:"website,omitempty"`
	Timezone         string         `gorm:"size:50;default:UTC" json:"timezone"`
	Language         string         `gorm:"size:10;default:en" json:"language"`
	Currency         string         `gorm:"size:3;default:USD" json:"currency"`
	Country          *string        `gorm:"size:2" json:"country,omitempty"`
	Industry         *string        `gorm:"size:50" json:"industry,omitempty"`
	CompanySize      *string        `gorm:"size:20" json:"company_size,omitempty"`
	BillingEmail     *string        `gorm:"size:255" json:"billing_email,omitempty"`
	Settings         types.JSON     `gorm:"type:jsonb;default:'{}'" json:"settings"`
	PlanID           string         `gorm:"size:50;default:free" json:"plan_id"`
	StripeCustomerID *string        `gorm:"size:255" json:"-"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations - loaded separately
	Members []Member `gorm:"foreignKey:WorkspaceID" json:"-"`
}

func (Workspace) TableName() string {
	return "workspaces"
}

// NewWorkspace creates a new workspace
func NewWorkspace(ownerID uuid.UUID, name, slug string) *Workspace {
	return &Workspace{
		ID:        uuid.New(),
		OwnerID:   ownerID,
		Name:      name,
		Slug:      slug,
		Timezone:  "UTC",
		Language:  "en",
		Currency:  "USD",
		PlanID:    PlanFree,
		Settings:  make(types.JSON),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// IsOwner checks if user is the workspace owner
func (w *Workspace) IsOwner(userID uuid.UUID) bool {
	return w.OwnerID == userID
}

// Update updates workspace details
func (w *Workspace) Update(name string, description, website, logoURL *string) {
	w.Name = name
	w.Description = description
	w.Website = website
	w.LogoURL = logoURL
	w.UpdatedAt = time.Now()
}

// UpdateSettings updates workspace settings
func (w *Workspace) UpdateSettings(settings types.JSON) {
	w.Settings = settings
	w.UpdatedAt = time.Now()
}

// UpdateLocale updates timezone and language
func (w *Workspace) UpdateLocale(timezone, language, currency string, country *string) {
	if timezone != "" {
		w.Timezone = timezone
	}
	if language != "" {
		w.Language = language
	}
	if currency != "" {
		w.Currency = currency
	}
	w.Country = country
	w.UpdatedAt = time.Now()
}

// SetPlan sets the workspace plan
func (w *Workspace) SetPlan(planID string) {
	w.PlanID = planID
	w.UpdatedAt = time.Now()
}

// SetStripeCustomer sets the Stripe customer ID
func (w *Workspace) SetStripeCustomer(customerID string) {
	w.StripeCustomerID = &customerID
	w.UpdatedAt = time.Now()
}

// TransferOwnership transfers workspace ownership to another user
func (w *Workspace) TransferOwnership(newOwnerID uuid.UUID) {
	w.OwnerID = newOwnerID
	w.UpdatedAt = time.Now()
}

// GetSetting retrieves a setting value
func (w *Workspace) GetSetting(key string) interface{} {
	if w.Settings == nil {
		return nil
	}
	return w.Settings[key]
}

// SetSetting sets a setting value
func (w *Workspace) SetSetting(key string, value interface{}) {
	if w.Settings == nil {
		w.Settings = make(types.JSON)
	}
	w.Settings[key] = value
	w.UpdatedAt = time.Now()
}

// Plan constants
const (
	PlanFree       = "free"
	PlanStarter    = "starter"
	PlanPro        = "pro"
	PlanEnterprise = "enterprise"
)
