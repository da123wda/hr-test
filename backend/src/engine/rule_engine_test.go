package engine

import (
	"hr-dynamic-form-engine/src/model"
	"strings"
	"testing"
)

func TestRuleEngine_CycleDetection(t *testing.T) {
	re := NewRuleEngine()

	fieldsWithCycle := []model.FieldDefinition{
		{
			Key:   "fieldA",
			Label: "字段A",
			Type:  "input",
			Rules: []model.FieldRule{
				{TargetKey: "fieldB", Action: "show", Expression: "fieldA == '1'"},
			},
		},
		{
			Key:   "fieldB",
			Label: "字段B",
			Type:  "input",
			Rules: []model.FieldRule{
				{TargetKey: "fieldC", Action: "show", Expression: "fieldB == '2'"},
			},
		},
		{
			Key:   "fieldC",
			Label: "字段C",
			Type:  "input",
			Rules: []model.FieldRule{
				{TargetKey: "fieldA", Action: "show", Expression: "fieldC == '3'"},
			},
		},
	}

	hasCycle, path := re.DetectCycles(fieldsWithCycle)
	if !hasCycle {
		t.Fatalf("预期检测到循环依赖闭环，但未检测出")
	}

	t.Logf("检测到循环依赖路径: %s", strings.Join(path, " -> "))
}

func TestRuleEngine_LinkageRules(t *testing.T) {
	re := NewRuleEngine()

	fields := []model.FieldDefinition{
		{
			Key:          "isFreshGraduate",
			Label:        "是否应届生",
			Type:         "radio",
			DefaultValue: "yes",
			Rules: []model.FieldRule{
				{TargetKey: "previousCompany", Action: "hide", Expression: "isFreshGraduate == 'yes'"},
				{TargetKey: "previousCompany", Action: "unsetRequired", Expression: "isFreshGraduate == 'yes'"},
			},
		},
		{
			Key:   "previousCompany",
			Label: "前公司名称",
			Type:  "input",
			Validation: &model.ValidationRules{
				Required: true,
			},
		},
		{
			Key:   "expectedSalary",
			Label: "期望薪资",
			Type:  "number",
			Rules: []model.FieldRule{
				{
					TargetKey:  "salaryAlert",
					Action:     "showAlert",
					Expression: "expectedSalary > 30000",
					Parameters: map[string]any{"text": "需要 HRBP 特别审批"},
				},
			},
		},
		{
			Key:   "salaryAlert",
			Label: "审批提示",
			Type:  "alert",
		},
	}

	data1 := map[string]interface{}{
		"isFreshGraduate": "yes",
		"expectedSalary":  35000,
	}
	states1, err := re.EvaluateRules(fields, data1)
	if err != nil {
		t.Fatalf("规则求值失败: %v", err)
	}

	if states1["previousCompany"].Visible != false {
		t.Errorf("应届生为是时，前公司名称字段应被隐藏")
	}
	if states1["previousCompany"].Required != false {
		t.Errorf("应届生为是时，前公司名称必填校验应被取消")
	}
	if states1["salaryAlert"].Alert != "需要 HRBP 特别审批" {
		t.Errorf("期望薪资 > 30000 时，应触发 HRBP 特别审批提示，实际得到: %s", states1["salaryAlert"].Alert)
	}

	data2 := map[string]interface{}{
		"isFreshGraduate": "no",
		"expectedSalary":  20000,
	}
	states2, err := re.EvaluateRules(fields, data2)
	if err != nil {
		t.Fatalf("规则求值失败: %v", err)
	}

	if states2["previousCompany"].Visible != true {
		t.Errorf("非应届生时，前公司名称字段应显示")
	}
	if states2["previousCompany"].Required != true {
		t.Errorf("非应届生时，前公司名称应为必填")
	}
	if states2["salaryAlert"].Alert != "" {
		t.Errorf("期望薪资 <= 30000 时，不应触发审批警示")
	}
}
