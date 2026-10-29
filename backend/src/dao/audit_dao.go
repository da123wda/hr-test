package dao

import (
	"hr-dynamic-form-engine/src/dto"
	"hr-dynamic-form-engine/src/model"

	"xorm.io/xorm"
)

type auditDao struct {
	conn *xorm.Engine
}

func AuditDbDefault() auditDao {
	return auditDao{conn: DBDefault()}
}

func (d auditDao) InsertOne(log *model.AuditLog) error {
	_, err := d.conn.InsertOne(log)
	return err
}

func (d auditDao) Query(targetType, targetId, formId string, page, pageSize int) (*dto.PageResult[model.AuditLog], error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 15
	}
	offset := (page - 1) * pageSize

	session := d.conn.Table("audit_log")
	if targetType != "" {
		session = session.Where("target_type = ?", targetType)
	}
	if targetId != "" {
		session = session.Where("target_id = ?", targetId)
	}
	if formId != "" {
		session = session.Where("form_id = ?", formId)
	}

	total, err := session.Count()
	if err != nil {
		return nil, err
	}

	var list []model.AuditLog
	qSession := d.conn.Table("audit_log")
	if targetType != "" {
		qSession = qSession.Where("target_type = ?", targetType)
	}
	if targetId != "" {
		qSession = qSession.Where("target_id = ?", targetId)
	}
	if formId != "" {
		qSession = qSession.Where("form_id = ?", formId)
	}

	err = qSession.OrderBy("created_at DESC").Limit(pageSize, offset).Find(&list)
	if err != nil {
		return nil, err
	}

	return &dto.PageResult[model.AuditLog]{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Items:    list,
	}, nil
}
