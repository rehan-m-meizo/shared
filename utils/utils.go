package utils

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// ToSnakeCase converts strings like "Branch Code", "branchCode", "BranchCode" → "branch_code"
func ToSnakeCase(str string) string {
	// Replace spaces, hyphens, dots etc. with underscore
	str = strings.TrimSpace(str)
	str = strings.ReplaceAll(str, "-", " ")
	str = strings.ReplaceAll(str, ".", " ")
	str = strings.ReplaceAll(str, "  ", " ") // optional: normalize multiple spaces
	str = strings.ReplaceAll(str, " ", "_")

	// Convert camelCase or PascalCase to snake_case
	re := regexp.MustCompile("([a-z0-9])([A-Z])")
	snake := re.ReplaceAllString(str, "${1}_${2}")

	return strings.ToLower(snake)
}

func PqQuoteIdentifier(name string) string {
	return "\"" + name + "\""
}

func MapToSQLType(fieldType string) string {
	switch fieldType {
	case "text", "ddl", "employee":
		return "VARCHAR(255)"
	case "number":
		return "INTEGER"
	case "date":
		return "DATE"
	case "boolean":
		return "BOOLEAN"
	default:
		return "TEXT"
	}
}

func Map[T any, R any](values []T, f func(T) R) []R {
	result := make([]R, len(values))
	for i, v := range values {
		result[i] = f(v)
	}
	return result
}

func SplitAndTrim(csv string) []string {
	parts := strings.Split(csv, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func ComputeSchemaHash(fields []interface{}) string {
	data, _ := json.Marshal(fields)
	hash := sha256.Sum256(data)
	return fmt.Sprintf("sha256:%x", hash)
}
