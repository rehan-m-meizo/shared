/**
 * @Author:
 * @Date: 2025-09-20 19:02:52
 * @LastEditors:
 * @LastEditTime: 2025-09-20 19:10:05
 * @FilePath: shared/pkgs/evaluate/evaluate.go
 * @Description: 这是默认设置，可以在设置》工具》File Description 中进行配置
 */
package evaluate

import (
	"fmt"
	"strings"

	"github.com/Knetic/govaluate"
)

type ConditionNode struct {
	All []ConditionNode `bson:"all,omitempty" json:"all,omitempty"` // AND
	Any []ConditionNode `bson:"any,omitempty" json:"any,omitempty"` // OR
	LHS string          `bson:"lhs,omitempty" json:"lhs,omitempty"` // e.g., "user.department"
	Op  string          `bson:"op,omitempty" json:"op,omitempty"`   // ==, !=, >, <, in, matches
	RHS interface{}     `bson:"rhs,omitempty" json:"rhs,omitempty"` // value or attribute reference
}

func conditionToExpression(node ConditionNode) string {
	if len(node.All) > 0 {
		parts := []string{}
		for _, child := range node.All {
			parts = append(parts, conditionToExpression(child))
		}
		return "(" + strings.Join(parts, " && ") + ")"
	}
	if len(node.Any) > 0 {
		var parts []string
		for _, child := range node.Any {
			parts = append(parts, conditionToExpression(child))
		}
		return "(" + strings.Join(parts, " || ") + ")"
	}
	// Leaf node: convert LHS, Op, RHS
	rhs := ""
	switch v := node.RHS.(type) {
	case string:
		rhs = fmt.Sprintf(`"%s"`, v)
	default:
		rhs = fmt.Sprintf(`%v`, v)
	}
	return fmt.Sprintf("%s %s %s", node.LHS, node.Op, rhs)
}

func Evaluate(condition ConditionNode, params map[string]interface{}) (bool, error) {
	exprStr := conditionToExpression(condition)

	expression, err := govaluate.NewEvaluableExpression(exprStr)

	if err != nil {
		return false, err
	}

	result, err := expression.Evaluate(params)

	if err != nil {
		return false, err
	}

	allowed, ok := result.(bool)

	if !ok {
		return false, fmt.Errorf("expected boolean result, got %T", result)
	}

	return allowed, nil
}
