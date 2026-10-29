package dto

import "hr-dynamic-form-engine/src/model"

type SubmitDataReq struct {
	SubmitterId   string                 `json:"submitterId"`
	SubmitterName string                 `json:"submitterName"`
	Data          map[string]interface{} `json:"data"`
}

type UpdateDataReq struct {
	OperatorId   string                 `json:"operatorId"`
	OperatorName string                 `json:"operatorName"`
	Data         map[string]interface{} `json:"data" binding:"required"`
}

type QuerySubmissionFilter struct {
	Page          int                    `form:"page"`
	PageSize      int                    `form:"pageSize"`
	SchemaVersion int                    `form:"schemaVersion"`
	Status        string                 `form:"status"`
	Keyword       string                 `form:"keyword"`
	Filters       map[string]interface{} `json:"filters"` // 动态 JSON 过滤 (如 {"experience": {">": 3}})
}

type PageResult[T any] struct {
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
	Items    []T   `json:"items"`
}

type SubmissionResp struct {
	SubmissionId  string `json:"submissionId"`
	SchemaVersion int    `json:"schemaVersion"`
	Status        string `json:"status"`
	CreatedAt     string `json:"createdAt"`
}

type SubmissionDetailResp struct {
	model.FormSubmission
}
