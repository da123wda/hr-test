package engine

import (
	"fmt"
	"hr-dynamic-form-engine/src/model"
	"regexp"
	"strconv"
	"strings"
)

type DynamicFieldState struct {
	Visible  bool   `json:"visible"`
	Required bool   `json:"required"`
	Alert    string `json:"alert"`
}

type RuleEngine struct {
	MaxEvaluationDepth int
}

func NewRuleEngine() *RuleEngine {
	return &RuleEngine{
		MaxEvaluationDepth: 20,
	}
}

// DetectCycles 依赖图环检测，判断字段间规则是否存在循环依赖
func (re *RuleEngine) DetectCycles(fields []model.FieldDefinition) (bool, []string) {
	adj := make(map[string][]string)
	allNodes := make(map[string]bool)

	for _, f := range fields {
		allNodes[f.Key] = true
		for _, rule := range f.Rules {
			allNodes[rule.TargetKey] = true
			sources := extractSourceKeys(rule.Expression, fields)
			if len(sources) == 0 {
				sources = []string{f.Key}
			}
			for _, src := range sources {
				adj[src] = append(adj[src], rule.TargetKey)
			}
		}
	}

	color := make(map[string]int)
	path := make([]string, 0)
	var cyclePath []string

	var dfs func(u string) bool
	dfs = func(u string) bool {
		color[u] = 1
		path = append(path, u)

		for _, v := range adj[u] {
			if color[v] == 1 {
				cycleStart := -1
				for i, n := range path {
					if n == v {
						cycleStart = i
						break
					}
				}
				if cycleStart != -1 {
					cyclePath = append([]string{}, path[cycleStart:]...)
					cyclePath = append(cyclePath, v)
				} else {
					cyclePath = []string{u, v}
				}
				return true
			}
			if color[v] == 0 {
				if dfs(v) {
					return true
				}
			}
		}

		color[u] = 2
		path = path[:len(path)-1]
		return false
	}

	for node := range allNodes {
		if color[node] == 0 {
			if dfs(node) {
				return true, cyclePath
			}
		}
	}

	return false, nil
}

// EvaluateRules 根据提交数据动态评估字段显隐、必填及提示状态
func (re *RuleEngine) EvaluateRules(fields []model.FieldDefinition, data map[string]interface{}) (map[string]*DynamicFieldState, error) {
	if hasCycle, path := re.DetectCycles(fields); hasCycle {
		return nil, fmt.Errorf("Rule Engine 检测到循环依赖闭环: %s", strings.Join(path, " -> "))
	}

	states := make(map[string]*DynamicFieldState)
	for _, f := range fields {
		req := false
		if f.Validation != nil && f.Validation.Required {
			req = true
		}
		states[f.Key] = &DynamicFieldState{
			Visible:  true,
			Required: req,
			Alert:    "",
		}
	}

	step := 0
	for _, f := range fields {
		for _, rule := range f.Rules {
			step++
			if step > re.MaxEvaluationDepth {
				return states, fmt.Errorf("Rule Engine 联动求值步数超过上限 %d", re.MaxEvaluationDepth)
			}

			if evaluateExpression(rule.Expression, data) {
				st, ok := states[rule.TargetKey]
				if !ok {
					st = &DynamicFieldState{Visible: true, Required: false}
					states[rule.TargetKey] = st
				}
				switch rule.Action {
				case "hide":
					st.Visible = false
				case "show":
					st.Visible = true
				case "setRequired":
					st.Required = true
				case "unsetRequired":
					st.Required = false
				case "showAlert":
					if txt, ok := rule.Parameters["text"].(string); ok {
						st.Alert = txt
					} else {
						st.Alert = "触发审批提醒"
					}
				}
			}
		}
	}

	return states, nil
}

func evaluateExpression(expr string, data map[string]interface{}) bool {
	expr = strings.TrimSpace(expr)
	if expr == "" || expr == "true" {
		return true
	}
	if expr == "false" {
		return false
	}

	if strings.Contains(expr, "&&") {
		for _, part := range strings.Split(expr, "&&") {
			if !evaluateExpression(part, data) {
				return false
			}
		}
		return true
	}
	if strings.Contains(expr, "||") {
		for _, part := range strings.Split(expr, "||") {
			if evaluateExpression(part, data) {
				return true
			}
		}
		return false
	}

	ops := []string{">=", "<=", "==", "!=", ">", "<"}
	var foundOp string
	for _, op := range ops {
		if strings.Contains(expr, op) {
			foundOp = op
			break
		}
	}

	if foundOp == "" {
		v, exists := data[expr]
		return exists && v != nil && v != false && v != "" && v != 0
	}

	parts := strings.SplitN(expr, foundOp, 2)
	leftKey := strings.TrimSpace(parts[0])
	rightClean := strings.Trim(strings.TrimSpace(parts[1]), `"'`)

	leftVal := data[leftKey]
	leftNum, isLeftNum := toFloat(leftVal)
	rightNum, errRight := strconv.ParseFloat(rightClean, 64)

	if isLeftNum && errRight == nil {
		switch foundOp {
		case "==":
			return leftNum == rightNum
		case "!=":
			return leftNum != rightNum
		case ">":
			return leftNum > rightNum
		case "<":
			return leftNum < rightNum
		case ">=":
			return leftNum >= rightNum
		case "<=":
			return leftNum <= rightNum
		}
	}

	leftStr := fmt.Sprintf("%v", leftVal)
	switch foundOp {
	case "==":
		return leftStr == rightClean
	case "!=":
		return leftStr != rightClean
	case ">":
		return leftStr > rightClean
	case "<":
		return leftStr < rightClean
	case ">=":
		return leftStr >= rightClean
	case "<=":
		return leftStr <= rightClean
	}

	return false
}

func extractSourceKeys(expr string, fields []model.FieldDefinition) []string {
	var keys []string
	reIdent := regexp.MustCompile(`[a-zA-Z_][a-zA-Z0-9_]*`)
	matches := reIdent.FindAllString(expr, -1)

	fm := make(map[string]bool)
	for _, f := range fields {
		fm[f.Key] = true
	}

	seen := make(map[string]bool)
	for _, m := range matches {
		if fm[m] && !seen[m] {
			seen[m] = true
			keys = append(keys, m)
		}
	}
	return keys
}

func toFloat(val interface{}) (float64, bool) {
	switch v := val.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case string:
		n, err := strconv.ParseFloat(v, 64)
		return n, err == nil
	}
	return 0, false
}
