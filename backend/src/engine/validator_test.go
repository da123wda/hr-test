package engine

import (
	"hr-dynamic-form-engine/src/model"
	"strings"
	"testing"
)

func TestFormValidator_Validation(t *testing.T) {
	validator := NewFormValidator()

	minLen := 2
	schema := &model.FormSchema{
		FormId: "test_form",
		Fields: []model.FieldDefinition{
			{
				Key:   "fullName",
				Label: "姓名",
				Type:  "input",
				Validation: &model.ValidationRules{
					Required:  true,
					MinLength: &minLen,
				},
			},
			{
				Key:   "email",
				Label: "邮箱",
				Type:  "input",
				Validation: &model.ValidationRules{
					Required: true,
					Email:    true,
				},
			},
			{
				Key:          "isFreshGraduate",
				Label:        "是否应届生",
				Type:         "radio",
				DefaultValue: "no",
				Rules: []model.FieldRule{
					{TargetKey: "previousCompany", Action: "hide", Expression: "isFreshGraduate == 'yes'"},
					{TargetKey: "previousCompany", Action: "unsetRequired", Expression: "isFreshGraduate == 'yes'"},
				},
			},
			{
				Key:   "previousCompany",
				Label: "前任公司",
				Type:  "input",
				Validation: &model.ValidationRules{
					Required: true,
				},
			},
			{
				Key:   "workHistory",
				Label: "工作履历",
				Type:  "repeater",
				ItemFields: []model.FieldDefinition{
					{
						Key:   "company",
						Label: "公司名",
						Type:  "input",
						Validation: &model.ValidationRules{
							Required: true,
						},
					},
				},
			},
			{
				Key:   "avatar",
				Label: "头像",
				Type:  "upload",
			},
		},
	}

	invalidData := map[string]interface{}{
		"fullName":        "张",
		"email":           "not-an-email",
		"isFreshGraduate": "no",
	}
	res1 := validator.Validate(schema, invalidData)
	if res1.IsValid {
		t.Fatalf("预期校验失败，但结果通过")
	}
	if !strings.Contains(res1.Errors["email"], "格式不正确") {
		t.Errorf("未能正确报错 email 格式错误: %v", res1.Errors["email"])
	}
	if res1.Errors["previousCompany"] == "" {
		t.Errorf("非应届生时未填 previousCompany 应报错")
	}

	validData := map[string]interface{}{
		"fullName":        "张三丰",
		"email":           "sanfeng@wudang.org",
		"isFreshGraduate": "yes",
		"workHistory": []interface{}{
			map[string]interface{}{"company": "武当山"},
		},
	}
	res2 := validator.Validate(schema, validData)
	if !res2.IsValid {
		t.Fatalf("预期校验通过，实际失败: %v", res2.Errors)
	}

	bigBinaryData := map[string]interface{}{
		"fullName":        "李四",
		"email":           "lisi@example.com",
		"isFreshGraduate": "yes",
		"avatar":          "data:image/png;base64," + strings.Repeat("A", 1000),
	}
	res3 := validator.Validate(schema, bigBinaryData)
	if res3.IsValid {
		t.Fatalf("文件字段直接提交大 Base64 数据应被拦截，实际未拦截")
	}
}
