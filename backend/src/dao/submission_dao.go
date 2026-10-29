package dao

import (
	"encoding/json"
	"errors"
	"fmt"
	"hr-dynamic-form-engine/src/dto"
	"hr-dynamic-form-engine/src/model"
	"reflect"
	"strconv"
	"strings"
	"time"

	"xorm.io/xorm"
)

type submissionDao struct {
	conn *xorm.Engine
}

func SubmissionDbDefault() submissionDao {
	return submissionDao{conn: DBDefault()}
}

// InsertOne 插入提交记录
func (d submissionDao) InsertOne(sub *model.FormSubmission) error {
	dataJSON, err := json.Marshal(sub.Data)
	if err != nil {
		return err
	}
	_, err = d.conn.Exec(
		`INSERT INTO form_submission (id, form_id, schema_version, data, status, submitter_id, submitter_name, created_at, updated_at)
		 VALUES (?, ?, ?, ?::jsonb, ?, ?, ?, NOW(), NOW())`,
		sub.Id, sub.FormId, sub.SchemaVersion, string(dataJSON), sub.Status, sub.SubmitterId, sub.SubmitterName,
	)
	return err
}

// GetById 获取单条提交记录
func (d submissionDao) GetById(formId, id string) (*model.FormSubmission, error) {
	sub := new(model.FormSubmission)
	has, err := d.conn.Where("form_id = ? AND id = ?", formId, id).Get(sub)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("记录未找到")
	}
	return sub, nil
}

// UpdateData 更新数据并自动记录审计日志
func (d submissionDao) UpdateData(sub *model.FormSubmission, operatorId, operatorName, clientIp string) error {
	oldSub, err := d.GetById(sub.FormId, sub.Id)
	if err != nil {
		return err
	}

	diff := calculateDataDiff(oldSub.Data, sub.Data)

	session := d.conn.NewSession()
	defer session.Close()
	if err = session.Begin(); err != nil {
		return err
	}

	dataJSON, err := json.Marshal(sub.Data)
	if err != nil {
		return err
	}

	_, err = session.Exec(
		`UPDATE form_submission SET data = ?::jsonb, updated_at = NOW() WHERE form_id = ? AND id = ?`,
		string(dataJSON), sub.FormId, sub.Id,
	)
	if err != nil {
		_ = session.Rollback()
		return err
	}

	// 记录审计日志
	if len(diff) > 0 {
		auditLog := &model.AuditLog{
			Action:       "UPDATE_SUBMISSION",
			TargetType:   "SUBMISSION",
			TargetId:     sub.Id,
			FormId:       sub.FormId,
			OperatorId:   operatorId,
			OperatorName: operatorName,
			ClientIp:     clientIp,
			Diff:         diff,
			CreatedAt:    time.Now(),
		}
		if _, err = session.InsertOne(auditLog); err != nil {
			_ = session.Rollback()
			return err
		}
	}

	return session.Commit()
}

// DeleteById 删除记录
func (d submissionDao) DeleteById(formId, id, operatorId, operatorName, clientIp string) error {
	session := d.conn.NewSession()
	defer session.Close()
	if err := session.Begin(); err != nil {
		return err
	}

	affected, err := session.Where("form_id = ? AND id = ?", formId, id).Delete(new(model.FormSubmission))
	if err != nil {
		_ = session.Rollback()
		return err
	}
	if affected == 0 {
		_ = session.Rollback()
		return errors.New("记录未找到")
	}

	auditLog := &model.AuditLog{
		Action:       "DELETE_SUBMISSION",
		TargetType:   "SUBMISSION",
		TargetId:     id,
		FormId:       formId,
		OperatorId:   operatorId,
		OperatorName: operatorName,
		ClientIp:     clientIp,
		CreatedAt:    time.Now(),
	}
	if _, err = session.InsertOne(auditLog); err != nil {
		_ = session.Rollback()
		return err
	}

	return session.Commit()
}

// QuerySubmissions 基于 PostgreSQL JSONB 的通用查询与动态字段筛选 (Q1: experience > 3)
func (d submissionDao) QuerySubmissions(formId string, filter *dto.QuerySubmissionFilter) (*dto.PageResult[model.FormSubmission], error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 10
	}

	var whereClauses []string
	var args []interface{}

	whereClauses = append(whereClauses, "form_id = ?")
	args = append(args, formId)

	if filter.SchemaVersion > 0 {
		whereClauses = append(whereClauses, "schema_version = ?")
		args = append(args, filter.SchemaVersion)
	}

	if filter.Status != "" {
		whereClauses = append(whereClauses, "status = ?")
		args = append(args, filter.Status)
	}

	if filter.Keyword != "" {
		whereClauses = append(whereClauses, "data::text ILIKE ?")
		args = append(args, "%"+filter.Keyword+"%")
	}

	// 动态 JSONB 字段解析与高级筛选
	for key, val := range filter.Filters {
		switch v := val.(type) {
		case map[string]interface{}:
			for op, targetVal := range v {
				sqlOp := "="
				switch op {
				case ">", "gt":
					sqlOp = ">"
				case ">=", "gte":
					sqlOp = ">="
				case "<", "lt":
					sqlOp = "<"
				case "<=", "lte":
					sqlOp = "<="
				case "!=", "ne":
					sqlOp = "!="
				}

				if num, ok := toFloat(targetVal); ok {
					whereClauses = append(whereClauses, fmt.Sprintf("(data->>'%s')::numeric %s ?", key, sqlOp))
					args = append(args, num)
				} else {
					whereClauses = append(whereClauses, fmt.Sprintf("data->>'%s' %s ?", key, sqlOp))
					args = append(args, fmt.Sprintf("%v", targetVal))
				}
			}
		default:
			whereClauses = append(whereClauses, fmt.Sprintf("data->>'%s' = ?", key))
			args = append(args, fmt.Sprintf("%v", val))
		}
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	// 1. 统计总数
	countSession := d.conn.Table("form_submission")
	if len(whereClauses) > 0 {
		countSession = countSession.Where(whereSQL, args...)
	}
	total, err := countSession.Count()
	if err != nil {
		return nil, err
	}

	// 2. 分页查询
	offset := (filter.Page - 1) * filter.PageSize
	var list []model.FormSubmission
	querySession := d.conn.Table("form_submission")
	if len(whereClauses) > 0 {
		querySession = querySession.Where(whereSQL, args...)
	}
	err = querySession.OrderBy("created_at DESC").Limit(filter.PageSize, offset).Find(&list)
	if err != nil {
		return nil, err
	}

	return &dto.PageResult[model.FormSubmission]{
		Total:    total,
		Page:     filter.Page,
		PageSize: filter.PageSize,
		Items:    list,
	}, nil
}

func calculateDataDiff(oldData, newData map[string]interface{}) map[string]model.DiffItem {
	diff := make(map[string]model.DiffItem)
	allKeys := make(map[string]bool)
	for k := range oldData {
		allKeys[k] = true
	}
	for k := range newData {
		allKeys[k] = true
	}

	for k := range allKeys {
		oldVal := oldData[k]
		newVal := newData[k]
		if !reflect.DeepEqual(oldVal, newVal) {
			diff[k] = model.DiffItem{
				Old: oldVal,
				New: newVal,
			}
		}
	}
	return diff
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
