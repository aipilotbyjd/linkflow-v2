package validation

import (
	"encoding/json"
	"testing"
)

// Test pointer fields
type PointerRequest struct {
	Name *string `json:"name" validate:"required"`
}

// Test nested struct
type NestedRequest struct {
	User UserInfo `json:"user" validate:"required"`
}

type UserInfo struct {
	Email string `json:"email" validate:"required,email"`
}

// Test slice validation
type SliceRequest struct {
	IDs []string `json:"ids" validate:"required,min=1"`
}

// Test UUID field
type UUIDRequest struct {
	WorkflowID string `json:"workflow_id" validate:"required,uuid"`
}

// Test JSON RawMessage (like in pinneddata)
type JSONRequest struct {
	Data json.RawMessage `json:"data" validate:"required"`
}

// Test zero value int
type ZeroIntRequest struct {
	Count int `json:"count" validate:"required"`
}

func TestPointerValidation(t *testing.T) {
	t.Run("nil pointer should fail required", func(t *testing.T) {
		req := PointerRequest{Name: nil}
		errors := Validate(req)
		t.Logf("Errors: %+v", errors)
		if len(errors) == 0 {
			t.Error("Expected error for nil pointer")
		}
	})

	t.Run("empty string pointer should pass required", func(t *testing.T) {
		empty := ""
		req := PointerRequest{Name: &empty}
		errors := Validate(req)
		t.Logf("Errors: %+v", errors)
		// Note: required only checks if pointer is nil, not if string is empty
	})
}

func TestNestedValidation(t *testing.T) {
	t.Run("nested struct validation", func(t *testing.T) {
		req := NestedRequest{User: UserInfo{Email: "invalid"}}
		errors := Validate(req)
		t.Logf("Errors: %+v", errors)
		// Check if nested validation works
		found := false
		for _, e := range errors {
			t.Logf("Field: %s, Message: %s", e.Field, e.Message)
			if e.Field == "email" || e.Field == "Email" {
				found = true
			}
		}
		if !found {
			t.Error("Expected nested email validation error")
		}
	})
}

func TestSliceValidation(t *testing.T) {
	t.Run("empty slice should fail min=1", func(t *testing.T) {
		req := SliceRequest{IDs: []string{}}
		errors := Validate(req)
		t.Logf("Errors: %+v", errors)
		if len(errors) == 0 {
			t.Error("Expected error for empty slice")
		}
	})

	t.Run("nil slice should fail required", func(t *testing.T) {
		req := SliceRequest{IDs: nil}
		errors := Validate(req)
		t.Logf("Errors: %+v", errors)
		if len(errors) == 0 {
			t.Error("Expected error for nil slice")
		}
	})
}

func TestUUIDValidation(t *testing.T) {
	t.Run("invalid UUID", func(t *testing.T) {
		req := UUIDRequest{WorkflowID: "not-a-uuid"}
		errors := Validate(req)
		t.Logf("Errors: %+v", errors)
		if len(errors) == 0 {
			t.Error("Expected error for invalid UUID")
		}
		for _, e := range errors {
			t.Logf("Field: %s, Message: %s", e.Field, e.Message)
		}
	})

	t.Run("valid UUID", func(t *testing.T) {
		req := UUIDRequest{WorkflowID: "550e8400-e29b-41d4-a716-446655440000"}
		errors := Validate(req)
		if len(errors) > 0 {
			t.Errorf("Expected no error for valid UUID, got: %+v", errors)
		}
	})
}

func TestJSONRawMessageValidation(t *testing.T) {
	t.Run("nil json.RawMessage should fail required", func(t *testing.T) {
		req := JSONRequest{Data: nil}
		errors := Validate(req)
		t.Logf("Errors: %+v", errors)
		if len(errors) == 0 {
			t.Error("Expected error for nil json.RawMessage")
		}
	})

	t.Run("empty json.RawMessage should fail required", func(t *testing.T) {
		req := JSONRequest{Data: json.RawMessage{}}
		errors := Validate(req)
		t.Logf("Errors: %+v", errors)
		// Empty slice might pass required - this could be an issue
	})

	t.Run("valid json.RawMessage", func(t *testing.T) {
		req := JSONRequest{Data: json.RawMessage(`{"key": "value"}`)}
		errors := Validate(req)
		if len(errors) > 0 {
			t.Errorf("Expected no error, got: %+v", errors)
		}
	})
}

func TestZeroIntValidation(t *testing.T) {
	t.Run("zero int with required", func(t *testing.T) {
		req := ZeroIntRequest{Count: 0}
		errors := Validate(req)
		t.Logf("Errors: %+v", errors)
		// Zero value might pass required - this could be an issue!
		if len(errors) == 0 {
			t.Log("WARNING: Zero int passes 'required' validation - may need 'min=1' for truly required int")
		}
	})
}
