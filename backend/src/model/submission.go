package model

import "time"

// FormSubmission 动态数据提交记录表
type FormSubmission struct {
	Id            string                 `json:"id" xorm:"varchar(64) pk 'id'"`
	FormId        string                 `json:"formId" xorm:"varchar(64) notnull index 'form_id'"`
	SchemaVersion int                    `json:"schemaVersion" xorm:"int notnull index 'schema_version'"` // 绑定历史版本 (Q2)
	Data          map[string]interface{} `json:"data" xorm:"jsonb notnull 'data'"`                      // PostgreSQL JSONB
	Status        string                 `json:"status" xorm:"varchar(32) notnull default 'submitted' 'status'"`
	SubmitterId   string                 `json:"submitterId" xorm:"varchar(64) 'submitter_id'"`
	SubmitterName string                 `json:"submitterName" xorm:"varchar(128) 'submitter_name'"`
	CreatedAt     time.Time              `json:"createdAt" xorm:"created 'created_at'"`
	UpdatedAt     time.Time              `json:"updatedAt" xorm:"updated 'updated_at'"`
}
