package types

import (
	"database/sql/driver"
	"fmt"
	"strings"
)

// Hstore represents the PostgreSQL hstore type
type Hstore map[string]string

// Value implements the driver.Valuer interface
func (h Hstore) Value() (driver.Value, error) {
	var parts []string
	for k, v := range h {
		parts = append(parts, fmt.Sprintf(`"%s"=>"%s"`, k, v))
	}
	return strings.Join(parts, ", "), nil
}

// Scan implements the sql.Scanner interface
func (h *Hstore) Scan(value interface{}) error {
	if value == nil {
		*h = nil
		return nil
	}
	s, ok := value.(string)
	if !ok {
		return fmt.Errorf("hstore Scan: expected string, got %T", value)
	}

	// Basic parsing for demonstration, robust parsing would be more complex
	parsedMap := make(map[string]string)
	pairs := strings.Split(s, ",")
	for _, pair := range pairs {
		parts := strings.Split(pair, "=>")
		if len(parts) == 2 {
			key := strings.Trim(parts[0], ` "`)
			val := strings.Trim(parts[1], ` "`)
			parsedMap[key] = val
		}
	}
	*h = parsedMap
	return nil
}

func HstoreToMapInteface(h Hstore) map[string]interface{} {
	result := make(map[string]interface{}, len(h))
	for k, v := range h {
		result[k] = v
	}
	return result
}

func MapToHstore(m map[string]interface{}) Hstore {
	result := make(Hstore, len(m))
	for k, v := range m {
		result[k] = fmt.Sprintf("%v", v)
	}

	return result
}
