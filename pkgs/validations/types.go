/**
 * @Author:
 * @Date: 2025-08-09 16:09:42
 * @LastEditors:
 * @LastEditTime: 2025-08-09 17:06:51
 * @FilePath: shared/pkgs/validations/types.go
 * @Description: 这是默认设置,可以在设置》工具》File Description中进行配置
 */
package validations

type Types struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

type Operators struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

func ValidationTypes() *Types {
	return &Types{}
}

func ListAllAvailableValidation() []Types {
	return []Types{
		{Key: "required", Label: "Required Field"},
		{Key: "string", Label: "String"},
		{Key: "email", Label: "Email Address"},
		{Key: "url", Label: "Url"},
		{Key: "max", Label: "Max Value"},
		{Key: "min", Label: "Min Value"},
		{Key: "len", Label: "Length"},
		{Key: "eq", Label: "Equal"},
		{Key: "ne", Label: "Not Equal"},
		{Key: "gte", Label: "Greater Than Equal"},
		{Key: "lte", Label: "Less Than Equal"},
		{Key: "contains", Label: "Contains"},
		{Key: "excludes", Label: "Excludes"},
		{Key: "uuid", Label: "Universally Unique Identifier"},
		{Key: "isbn", Label: "International Standard Book Number"},
		{Key: "isbn10", Label: "International Standard Book Number 10"},
		{Key: "isbn13", Label: "International Standard Book Number 13"},
		{Key: "mac", Label: "Media Access Control"},
		{Key: "ip", Label: "Internet Protocol"},
		{Key: "ipv4", Label: "Internet Protocol Version 4"},
		{Key: "ipv6", Label: "Internet Protocol Version 6"},
		{Key: "ascii", Label: "American Standard Code for Information Interchange"},
		{Key: "unique", Label: "Unique Field"},
		{Key: "numeric", Label: "Numeric"},
		{Key: "alpha", Label: "Alpha"},
		{Key: "alphanum", Label: "Alpha Numeric"},
		{Key: "pan", Label: "Pan Card Number"},
		{Key: "tan", Label: "Tan Card Number"},
	}
}

func GetAllRuleTypes() []Types {
	return []Types{
		{Key: "show_if", Label: "Show if matched"},
		{Key: "hide_if", Label: "Hide if matched"},
		{Key: "enable_if", Label: "Enable if matched"},
		{Key: "populate_from", Label: "Populate from another source"},
	}
}

func GetAllOperators() []Operators {
	return []Operators{
		{Key: "equals", Label: "Equals", Description: "Field value must equal specified value"},
		{Key: "not_equals", Label: "Does Not Equal", Description: "Field value must NOT equal specified value"},
		{Key: "greater_than", Label: "Greater Than", Description: "Field value must be greater than specified value"},
		{Key: "less_than", Label: "Less Than", Description: "Field value must be less than specified value"},
		{Key: "contains", Label: "Contains", Description: "Field value must contain specified value"},
		{Key: "not_contains", Label: "Does Not Contain", Description: "Field value must NOT contain specified value"},
		{Key: "matches_regex", Label: "Matches Pattern", Description: "Field value must match specified pattern"},
	}
}
