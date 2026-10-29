package engine

import (
	"fmt"
	"hr-dynamic-form-engine/src/model"
	"net/mail"
	"regexp"
	"strings"
)

type ValidationResult struct {
	IsValid bool              `json:"isValid"`
	Errors  map[string]string `json:"errors"`
}

type FormValidator struct {
	ruleEngine *RuleEngine
}

func NewFormValidator() *FormValidator {
	return &FormValidator{
		ruleEngine: NewRuleEngine(),
	}
}

// Validate 校验提交的数据是否符合对应 Schema 定义与联动条件
func (fv *FormValidator) Validate(schema *model.FormSchema, payload map[string]interface{}) *ValidationResult {
	res := &ValidationResult{
		IsValid: true,
		Errors:  make(map[string]string),
	}

	dynamicStates, err := fv.ruleEngine.EvaluateRules(schema.Fields, payload)
	if err != nil {
		res.IsValid = false
		res.Errors["_schema_rule_error"] = err.Error()
		return res
	}

	for _, field := range schema.Fields {
		if field.Deprecated {
			continue
		}

		state := dynamicStates[field.Key]
		if state != nil && !state.Visible {
			continue
		}

		isRequired := false
		if field.Validation != nil && field.Validation.Required {
			isRequired = true
		}
		if state != nil {
			isRequired = state.Required
		}

		val, exists := payload[field.Key]
		fv.validateSingleField(field.Key, field, val, exists, isRequired, res)
	}

	if len(res.Errors) > 0 {
		res.IsValid = false
	}
	return res
}

func (fv *FormValidator) validateSingleField(
	fieldPath string,
	field model.FieldDefinition,
	val interface{},
	exists bool,
	isRequired bool,
	res *ValidationResult,
) {
	isEmpty := !exists || val == nil || val == ""

	if isRequired && isEmpty {
		msg := fmt.Sprintf("%s 为必填项", field.Label)
		if field.Validation != nil && field.Validation.CustomMessage != "" {
			msg = field.Validation.CustomMessage
		}
		res.Errors[fieldPath] = msg
		return
	}

	if isEmpty {
		return
	}

	switch field.Type {
	case "input", "textarea":
		strVal, ok := val.(string)
		if !ok {
			res.Errors[fieldPath] = fmt.Sprintf("%s 必须为文本类型", field.Label)
			return
		}

		// 检查大文本/二进制字符串，避免将大二进制直接塞入 JSON
		if strings.HasPrefix(strVal, "data:") || len(strVal) > 100000 {
			res.Errors[fieldPath] = "文件类型字段禁止直接存储二进制大文本，请通过文件接口上传后使用"
			return
		}

		if field.Validation != nil && field.Validation.Regex != "" {
			matched, err := regexp.MatchString(field.Validation.Regex, strVal)
			if err != nil || !matched {
				res.Errors[fieldPath] = fmt.Sprintf("字段 %s 格式不正确", field.Label)
				return
			}
		}

		if field.Validation != nil && field.Validation.Email {
			if !isValidEmail(strVal) {
				res.Errors[fieldPath] = fmt.Sprintf("字段 %s 格式不正确", field.Label)
				return
			}
		}

		if field.Validation != nil {
			length := len([]rune(strVal))
			if field.Validation.MinLength != nil && length < *field.Validation.MinLength {
				res.Errors[fieldPath] = fmt.Sprintf("%s 最小长度为 %d 个字符", field.Label, *field.Validation.MinLength)
				return
			}
			if field.Validation.MaxLength != nil && length > *field.Validation.MaxLength {
				res.Errors[fieldPath] = fmt.Sprintf("%s 最大长度为 %d 个字符", field.Label, *field.Validation.MaxLength)
				return
			}
		}

	case "number":
		num, ok := toFloat(val)
		if !ok {
			res.Errors[fieldPath] = fmt.Sprintf("%s 必须为数值类型", field.Label)
			return
		}
		if field.Validation != nil {
			if field.Validation.Min != nil && num < *field.Validation.Min {
				res.Errors[fieldPath] = fmt.Sprintf("%s 不能小于 %v", field.Label, *field.Validation.Min)
				return
			}
			if field.Validation.Max != nil && num > *field.Validation.Max {
				res.Errors[fieldPath] = fmt.Sprintf("%s 不能大于 %v", field.Label, *field.Validation.Max)
				return
			}
		}

	case "select", "radio":
		if len(field.Options) > 0 {
			validOption := false
			strVal := fmt.Sprintf("%v", val)
			for _, opt := range field.Options {
				if fmt.Sprintf("%v", opt.Value) == strVal {
					validOption = true
					break
				}
			}
			if !validOption {
				res.Errors[fieldPath] = fmt.Sprintf("%s 的选择值不合法", field.Label)
				return
			}
		}

	case "upload":
		switch v := val.(type) {
		case map[string]interface{}:
			if v["fileId"] == nil || v["fileId"] == "" {
				res.Errors[fieldPath] = fmt.Sprintf("%s 上传文件信息缺失 fileId", field.Label)
				return
			}
		case string:
			if strings.HasPrefix(v, "data:") || len(v) > 500 {
				res.Errors[fieldPath] = "文件字段禁止直接存储二进制大数据"
				return
			}
		default:
			res.Errors[fieldPath] = fmt.Sprintf("%s 必须为文件元数据对象", field.Label)
			return
		}

	case "repeater":
		items, ok := val.([]interface{})
		if !ok {
			res.Errors[fieldPath] = fmt.Sprintf("%s 必须为列表数据", field.Label)
			return
		}
		if field.Validation != nil {
			if field.Validation.MinLength != nil && len(items) < *field.Validation.MinLength {
				res.Errors[fieldPath] = fmt.Sprintf("%s 至少需要 %d 项记录", field.Label, *field.Validation.MinLength)
				return
			}
			if field.Validation.MaxLength != nil && len(items) > *field.Validation.MaxLength {
				res.Errors[fieldPath] = fmt.Sprintf("%s 最多允许 %d 项记录", field.Label, *field.Validation.MaxLength)
				return
			}
		}
		for idx, itemRaw := range items {
			itemMap, isMap := itemRaw.(map[string]interface{})
			if !isMap {
				res.Errors[fmt.Sprintf("%s[%d]", fieldPath, idx)] = "数组项必须为对象结构"
				continue
			}
			for _, subField := range field.ItemFields {
				subPath := fmt.Sprintf("%s[%d].%s", fieldPath, idx, subField.Key)
				subVal, subExists := itemMap[subField.Key]
				subReq := subField.Validation != nil && subField.Validation.Required
				fv.validateSingleField(subPath, subField, subVal, subExists, subReq, res)
			}
		}

	case "object":
		objMap, ok := val.(map[string]interface{})
		if !ok {
			res.Errors[fieldPath] = fmt.Sprintf("%s 必须为对象类型", field.Label)
			return
		}
		for _, child := range field.Children {
			childPath := fmt.Sprintf("%s.%s", fieldPath, child.Key)
			childVal, childExists := objMap[child.Key]
			childReq := child.Validation != nil && child.Validation.Required
			fv.validateSingleField(childPath, child, childVal, childExists, childReq, res)
		}
	}
}

func isValidEmail(email string) bool {
	addr, err := mail.ParseAddress(email)
	return err == nil && addr.Address == email
}
