package integrations

import (
	"github.com/google/uuid"
)

func getString(config map[string]interface{}, key, defaultVal string) string {
	if v, ok := config[key].(string); ok {
		return v
	}
	return defaultVal
}

func getInt(config map[string]interface{}, key string, defaultVal int) int {
	if v, ok := config[key]; ok {
		switch val := v.(type) {
		case int:
			return val
		case int64:
			return int(val)
		case float64:
			return int(val)
		}
	}
	return defaultVal
}

func getFloat(config map[string]interface{}, key string, defaultVal float64) float64 {
	if v, ok := config[key]; ok {
		switch val := v.(type) {
		case float64:
			return val
		case int:
			return float64(val)
		case int64:
			return float64(val)
		}
	}
	return defaultVal
}

func getBool(config map[string]interface{}, key string, defaultVal bool) bool {
	if v, ok := config[key].(bool); ok {
		return v
	}
	return defaultVal
}

func parseUUID(s string) uuid.UUID {
	id, _ := uuid.Parse(s)
	return id
}
