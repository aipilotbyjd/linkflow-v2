package user

import (
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
	"gorm.io/gorm"
)

// User entity (aggregate root)
type User struct {
	ID                      uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Email                   string         `gorm:"uniqueIndex;size:255;not null" json:"email"`
	Username                *string        `gorm:"uniqueIndex;size:100" json:"username,omitempty"`
	PasswordHash            string         `gorm:"size:255" json:"-"`
	FirstName               string         `gorm:"size:100" json:"first_name"`
	LastName                string         `gorm:"size:100" json:"last_name"`
	AvatarURL               *string        `gorm:"size:500" json:"avatar_url,omitempty"`
	Phone                   *string        `gorm:"size:20" json:"phone,omitempty"`
	Bio                     *string        `gorm:"type:text" json:"bio,omitempty"`
	JobTitle                *string        `gorm:"size:100" json:"job_title,omitempty"`
	Company                 *string        `gorm:"size:100" json:"company,omitempty"`
	Timezone                string         `gorm:"size:50;default:UTC" json:"timezone"`
	Language                string         `gorm:"size:10;default:en" json:"language"`
	DateFormat              string         `gorm:"size:20;default:MM/DD/YYYY" json:"date_format"`
	TimeFormat              string         `gorm:"size:5;default:12h" json:"time_format"`
	Theme                   string         `gorm:"size:10;default:system" json:"theme"`
	NotificationPreferences types.JSON     `gorm:"type:jsonb;default:'{}'" json:"notification_preferences"`
	Status                  Status         `gorm:"size:20;default:active;index" json:"status"`
	EmailVerified           bool           `gorm:"default:false" json:"email_verified"`
	MFAEnabled              bool           `gorm:"default:false" json:"mfa_enabled"`
	MFASecret               *string        `gorm:"size:255" json:"-"`
	LastLoginAt             *time.Time     `json:"last_login_at,omitempty"`
	LoginCount              int            `gorm:"default:0" json:"login_count"`
	FailedLogins            int            `gorm:"default:0" json:"-"`
	LockedUntil             *time.Time     `json:"-"`
	CreatedAt               time.Time      `json:"created_at"`
	UpdatedAt               time.Time      `json:"updated_at"`
	DeletedAt               gorm.DeletedAt `gorm:"index" json:"-"`
}

func (User) TableName() string {
	return "users"
}

// NewUser creates a new user entity
func NewUser(email, passwordHash, firstName, lastName string) (*User, error) {
	if email == "" {
		return nil, ErrEmailRequired
	}
	if passwordHash == "" {
		return nil, ErrPasswordHashRequired
	}
	if firstName == "" {
		return nil, ErrFirstNameRequired
	}
	if len(firstName) > 100 {
		return nil, ErrFirstNameTooLong
	}
	if lastName == "" {
		return nil, ErrLastNameRequired
	}
	if len(lastName) > 100 {
		return nil, ErrLastNameTooLong
	}

	return &User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: passwordHash,
		FirstName:    firstName,
		LastName:     lastName,
		Status:       StatusActive,
		Timezone:     "UTC",
		Language:     "en",
		DateFormat:   "MM/DD/YYYY",
		TimeFormat:   "12h",
		Theme:        "system",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}, nil
}

// FullName returns the user's full name
func (u *User) FullName() string {
	return u.FirstName + " " + u.LastName
}

// IsActive checks if user is active
func (u *User) IsActive() bool {
	return u.Status == StatusActive
}

// IsLocked checks if the user account is locked
func (u *User) IsLocked() bool {
	if u.LockedUntil == nil {
		return false
	}
	return u.LockedUntil.After(time.Now())
}

// CanLogin checks if user can perform login
func (u *User) CanLogin() bool {
	return u.IsActive() && !u.IsLocked()
}

// IncrementLoginCount increments login count and resets failed logins
func (u *User) IncrementLoginCount() {
	u.LoginCount++
	u.FailedLogins = 0
	now := time.Now()
	u.LastLoginAt = &now
	u.UpdatedAt = time.Now()
}

// IncrementFailedLogin increments failed login count
func (u *User) IncrementFailedLogin() {
	u.FailedLogins++
	u.UpdatedAt = time.Now()
}

// LockAccount locks the user account until specified time
func (u *User) LockAccount(until time.Time) {
	u.LockedUntil = &until
	u.UpdatedAt = time.Now()
}

// UnlockAccount unlocks the user account
func (u *User) UnlockAccount() {
	u.LockedUntil = nil
	u.FailedLogins = 0
	u.UpdatedAt = time.Now()
}

// VerifyEmail marks email as verified
func (u *User) VerifyEmail() {
	u.EmailVerified = true
	u.UpdatedAt = time.Now()
}

// EnableMFA enables MFA with the given secret
func (u *User) EnableMFA(secret string) {
	u.MFAEnabled = true
	u.MFASecret = &secret
	u.UpdatedAt = time.Now()
}

// DisableMFA disables MFA
func (u *User) DisableMFA() {
	u.MFAEnabled = false
	u.MFASecret = nil
	u.UpdatedAt = time.Now()
}

// UpdatePassword updates the password hash
func (u *User) UpdatePassword(hash string) {
	u.PasswordHash = hash
	u.UpdatedAt = time.Now()
}

// UpdateProfile updates user profile information
func (u *User) UpdateProfile(firstName, lastName string, phone, bio, jobTitle, company *string) {
	u.FirstName = firstName
	u.LastName = lastName
	u.Phone = phone
	u.Bio = bio
	u.JobTitle = jobTitle
	u.Company = company
	u.UpdatedAt = time.Now()
}

// UpdatePreferences updates user preferences
func (u *User) UpdatePreferences(timezone, language, dateFormat, timeFormat, theme string) {
	if timezone != "" {
		u.Timezone = timezone
	}
	if language != "" {
		u.Language = language
	}
	if dateFormat != "" {
		u.DateFormat = dateFormat
	}
	if timeFormat != "" {
		u.TimeFormat = timeFormat
	}
	if theme != "" {
		u.Theme = theme
	}
	u.UpdatedAt = time.Now()
}

// Suspend suspends the user account
func (u *User) Suspend() {
	u.Status = StatusSuspended
	u.UpdatedAt = time.Now()
}

// Activate activates the user account
func (u *User) Activate() {
	u.Status = StatusActive
	u.UpdatedAt = time.Now()
}
