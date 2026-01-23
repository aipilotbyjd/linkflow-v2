package validation

import (
	"testing"
)

type TestRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Code     string `json:"code" validate:"required,len=6"`
	Score    int    `json:"score" validate:"required,min=1,max=5"`
	Role     string `json:"role" validate:"required,oneof=admin member viewer"`
}

func TestValidation(t *testing.T) {
	tests := []struct {
		name     string
		req      TestRequest
		expected map[string]string // field -> expected message substring
	}{
		{
			name: "Empty struct",
			req:  TestRequest{},
			expected: map[string]string{
				"email":    "Email is required",
				"password": "Password is required",
				"code":     "Code is required",
				"role":     "Role is required",
			},
		},
		{
			name: "Invalid email",
			req:  TestRequest{Email: "invalid", Password: "12345678", Code: "123456", Score: 3, Role: "admin"},
			expected: map[string]string{
				"email": "valid email",
			},
		},
		{
			name: "Short password",
			req:  TestRequest{Email: "test@test.com", Password: "123", Code: "123456", Score: 3, Role: "admin"},
			expected: map[string]string{
				"password": "at least 8",
			},
		},
		{
			name: "Wrong code length",
			req:  TestRequest{Email: "test@test.com", Password: "12345678", Code: "12345", Score: 3, Role: "admin"},
			expected: map[string]string{
				"code": "exactly 6",
			},
		},
		{
			name: "Score out of range",
			req:  TestRequest{Email: "test@test.com", Password: "12345678", Code: "123456", Score: 10, Role: "admin"},
			expected: map[string]string{
				"score": "at most 5",
			},
		},
		{
			name: "Invalid role",
			req:  TestRequest{Email: "test@test.com", Password: "12345678", Code: "123456", Score: 3, Role: "superadmin"},
			expected: map[string]string{
				"role": "one of",
			},
		},
		{
			name:     "Valid request",
			req:      TestRequest{Email: "test@test.com", Password: "12345678", Code: "123456", Score: 3, Role: "admin"},
			expected: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := Validate(tt.req)

			t.Logf("Test: %s", tt.name)
			for _, e := range errors {
				t.Logf("  Field: %s, Message: %s", e.Field, e.Message)
			}

			if len(tt.expected) == 0 && len(errors) > 0 {
				t.Errorf("Expected no errors but got %d", len(errors))
			}

			for field, substr := range tt.expected {
				found := false
				for _, e := range errors {
					if e.Field == field {
						found = true
						if !contains(e.Message, substr) {
							t.Errorf("Field %s: expected message containing '%s', got '%s'", field, substr, e.Message)
						}
					}
				}
				if !found {
					t.Errorf("Expected error for field %s but not found", field)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
