package types

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/lib/pq"
)

// JSON type for JSONB columns
type JSON map[string]interface{}

func (j JSON) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan JSON: not a byte slice")
	}
	return json.Unmarshal(bytes, j)
}

func (j JSON) Get(key string) interface{} {
	if j == nil {
		return nil
	}
	return j[key]
}

func (j JSON) GetString(key string) string {
	if j == nil {
		return ""
	}
	if v, ok := j[key].(string); ok {
		return v
	}
	return ""
}

func (j JSON) GetInt(key string) int {
	if j == nil {
		return 0
	}
	switch v := j[key].(type) {
	case int:
		return v
	case float64:
		return int(v)
	case int64:
		return int(v)
	default:
		return 0
	}
}

func (j JSON) GetBool(key string) bool {
	if j == nil {
		return false
	}
	if v, ok := j[key].(bool); ok {
		return v
	}
	return false
}

// JSONArray type for JSONB array columns
type JSONArray []interface{}

func (j JSONArray) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

func (j *JSONArray) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan JSONArray: not a byte slice")
	}
	return json.Unmarshal(bytes, j)
}

func (j JSONArray) Len() int {
	return len(j)
}

// StringArray type for text[] columns (PostgreSQL array)
type StringArray = pq.StringArray

// RawJSON stores raw JSON without parsing
type RawJSON []byte

func (r RawJSON) Value() (driver.Value, error) {
	if r == nil {
		return nil, nil
	}
	return []byte(r), nil
}

func (r *RawJSON) Scan(value interface{}) error {
	if value == nil {
		*r = nil
		return nil
	}
	switch v := value.(type) {
	case []byte:
		*r = v
		return nil
	case string:
		*r = []byte(v)
		return nil
	default:
		return errors.New("failed to scan RawJSON: unsupported type")
	}
}

func (r RawJSON) MarshalJSON() ([]byte, error) {
	if r == nil {
		return []byte("null"), nil
	}
	return r, nil
}

func (r *RawJSON) UnmarshalJSON(data []byte) error {
	if r == nil {
		return errors.New("RawJSON: UnmarshalJSON on nil pointer")
	}
	*r = append((*r)[0:0], data...)
	return nil
}
