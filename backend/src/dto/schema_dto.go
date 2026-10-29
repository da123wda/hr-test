package dto

import "hr-dynamic-form-engine/src/model"

type SaveSchemaReq struct {
	FormId          string                  `json:"formId" binding:"required"`
	Title           string                  `json:"title" binding:"required"`
	Description     string                  `json:"description"`
	ExpectedVersion int                     `json:"expectedVersion"` // Q5: 乐观锁预期版本号
	Fields          []model.FieldDefinition `json:"fields" binding:"required"`
	Migrations      map[string]string       `json:"migrations,omitempty"`
}

type PublishSchemaReq struct {
	ExpectedVersion int `json:"expectedVersion"`
}

type SchemaResp struct {
	model.FormSchema
	Etag string `json:"etag"`
}
