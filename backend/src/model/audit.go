package model

import "time"

// DiffItem 字段级变更项 (Q6)
type DiffItem struct {
	Old any `json:"old"`
	New any `json:"new"`
}

// AuditLog 审计日志表
type AuditLog struct {
	Id           int64               `json:"id" xorm:"pk autoincr 'id'"`
	Action       string              `json:"action" xorm:"varchar(64) notnull 'action'"`
	TargetType   string              `json:"targetType" xorm:"varchar(64) notnull 'target_type'"`
	TargetId     string              `json:"targetId" xorm:"varchar(128) notnull 'target_id'"`
	FormId       string              `json:"formId" xorm:"varchar(64) notnull index 'form_id'"`
	OperatorId   string              `json:"operatorId" xorm:"varchar(64) 'operator_id'"`
	OperatorName string              `json:"operatorName" xorm:"varchar(128) 'operator_name'"`
	ClientIp     string              `json:"clientIp" xorm:"varchar(64) 'client_ip'"`
	Diff         map[string]DiffItem `json:"diff" xorm:"jsonb 'diff'"` // 记录差异: {"salary": {"old": 30000, "new": 35000}}
	CreatedAt    time.Time           `json:"createdAt" xorm:"created 'created_at'"`
}
