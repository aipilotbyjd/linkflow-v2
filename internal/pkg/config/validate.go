package config

import (
	"errors"
	"fmt"
	"net/url"
)

// Validate validates the configuration and returns any errors
func (c *Config) Validate() error {
	var errs []error

	// JWT validation (critical)
	if c.JWT.Secret == "" {
		errs = append(errs, errors.New("jwt.secret is required"))
	} else if len(c.JWT.Secret) < 32 {
		errs = append(errs, errors.New("jwt.secret must be at least 32 characters"))
	}
	if c.JWT.Secret == "change-me-in-production" && c.App.Environment == "production" {
		errs = append(errs, errors.New("jwt.secret must be changed in production"))
	}

	// Database validation
	if c.Database.Host == "" {
		errs = append(errs, errors.New("database.host is required"))
	}
	if c.Database.Port <= 0 || c.Database.Port > 65535 {
		errs = append(errs, errors.New("database.port must be between 1 and 65535"))
	}
	if c.Database.User == "" {
		errs = append(errs, errors.New("database.user is required"))
	}
	if c.Database.Name == "" {
		errs = append(errs, errors.New("database.name is required"))
	}

	// Redis validation
	if c.Redis.Host == "" {
		errs = append(errs, errors.New("redis.host is required"))
	}
	if c.Redis.Port <= 0 || c.Redis.Port > 65535 {
		errs = append(errs, errors.New("redis.port must be between 1 and 65535"))
	}

	// Server validation
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		errs = append(errs, errors.New("server.port must be between 1 and 65535"))
	}

	// URL validation
	if c.App.URL != "" {
		if _, err := url.Parse(c.App.URL); err != nil {
			errs = append(errs, fmt.Errorf("app.url is invalid: %w", err))
		}
	}
	if c.App.FrontendURL != "" {
		if _, err := url.Parse(c.App.FrontendURL); err != nil {
			errs = append(errs, fmt.Errorf("app.frontend_url is invalid: %w", err))
		}
	}

	// Production-specific validations
	if c.App.Environment == "production" {
		if c.App.Debug {
			errs = append(errs, errors.New("app.debug should be false in production"))
		}
		if c.Database.SSLMode == "disable" {
			errs = append(errs, errors.New("database.sslmode should not be 'disable' in production"))
		}
	}

	if len(errs) > 0 {
		return &ConfigValidationError{Errors: errs}
	}
	return nil
}

// ValidateRequired validates only the critical configuration needed to start
func (c *Config) ValidateRequired() error {
	var errs []error

	if c.JWT.Secret == "" {
		errs = append(errs, errors.New("jwt.secret is required"))
	}
	if c.Database.Host == "" {
		errs = append(errs, errors.New("database.host is required"))
	}
	if c.Redis.Host == "" {
		errs = append(errs, errors.New("redis.host is required"))
	}

	if len(errs) > 0 {
		return &ConfigValidationError{Errors: errs}
	}
	return nil
}

// ConfigValidationError contains multiple validation errors
type ConfigValidationError struct {
	Errors []error
}

func (e *ConfigValidationError) Error() string {
	if len(e.Errors) == 1 {
		return fmt.Sprintf("config validation error: %s", e.Errors[0].Error())
	}
	msg := fmt.Sprintf("config validation errors (%d):", len(e.Errors))
	for _, err := range e.Errors {
		msg += "\n  - " + err.Error()
	}
	return msg
}

// IsProduction returns true if running in production environment
func (c *Config) IsProduction() bool {
	return c.App.Environment == "production"
}

// IsDevelopment returns true if running in development environment
func (c *Config) IsDevelopment() bool {
	return c.App.Environment == "development" || c.App.Environment == ""
}
