package service

import (
	"fmt"
	"hr-dynamic-form-engine/src/cbase"
	"hr-dynamic-form-engine/src/dao"
	"hr-dynamic-form-engine/src/dto"
	"hr-dynamic-form-engine/src/engine"
	"hr-dynamic-form-engine/src/model"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type DynamicServer struct {
	UserId   string
	UserName string
}

func (s DynamicServer) SubmitService(ctx *gin.Context, formId string, args *dto.SubmitDataReq) (*dto.SubmissionResp, *engine.ValidationResult) {
	schema, err := dao.FormSchemaDbDefault().GetByFormId(formId)
	if err != nil {
		cbase.PanicMsg(fmt.Sprintf("表单 [%s] 不存在", formId))
	}
	if schema.PublishedVersion == 0 {
		cbase.PanicMsg("该表单尚未发布，暂不可提交数据")
	}

	// 提取当前正式发布的版本快照，严格基于已发布版本进行校验
	snap, err := dao.FormSchemaDbDefault().GetVersion(formId, schema.PublishedVersion)
	if err == nil && snap != nil {
		schema.Fields = snap.Fields
		schema.Title = snap.Title
	}

	validator := engine.NewFormValidator()
	valRes := validator.Validate(schema, args.Data)
	if !valRes.IsValid {
		return nil, valRes
	}

	sub := &model.FormSubmission{
		Id:            "sub_" + uuid.New().String()[:8],
		FormId:        formId,
		SchemaVersion: schema.PublishedVersion,
		Data:          args.Data,
		Status:        "submitted",
		SubmitterId:   args.SubmitterId,
		SubmitterName: args.SubmitterName,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := dao.SubmissionDbDefault().InsertOne(sub); err != nil {
		cbase.PanicMsg("数据持久化失败: " + err.Error())
	}

	_ = dao.AuditDbDefault().InsertOne(&model.AuditLog{
		Action:       "CREATE_SUBMISSION",
		TargetType:   "SUBMISSION",
		TargetId:     sub.Id,
		FormId:       formId,
		OperatorId:   args.SubmitterId,
		OperatorName: args.SubmitterName,
		ClientIp:     ctx.ClientIP(),
	})

	return &dto.SubmissionResp{
		SubmissionId:  sub.Id,
		SchemaVersion: sub.SchemaVersion,
		Status:        sub.Status,
		CreatedAt:     sub.CreatedAt.Format(time.RFC3339),
	}, valRes
}

func (DynamicServer) QueryService(ctx *gin.Context, formId string, filter *dto.QuerySubmissionFilter) *dto.PageResult[model.FormSubmission] {
	res, err := dao.SubmissionDbDefault().QuerySubmissions(formId, filter)
	if err != nil {
		cbase.PanicMsg(err.Error())
	}
	return res
}

func (DynamicServer) GetDetailService(formId, id string) *model.FormSubmission {
	sub, err := dao.SubmissionDbDefault().GetById(formId, id)
	if err != nil {
		cbase.PanicMsg(err.Error())
	}
	return sub
}

func (DynamicServer) UpdateService(ctx *gin.Context, formId, id string, args *dto.UpdateDataReq) (*model.FormSubmission, *engine.ValidationResult) {
	sub, err := dao.SubmissionDbDefault().GetById(formId, id)
	if err != nil {
		cbase.PanicMsg(err.Error())
	}

	schemaVersion, err := dao.FormSchemaDbDefault().GetVersion(formId, sub.SchemaVersion)
	if err == nil {
		validator := engine.NewFormValidator()
		valRes := validator.Validate(&model.FormSchema{
			FormId: formId,
			Fields: schemaVersion.Fields,
		}, args.Data)
		if !valRes.IsValid {
			return nil, valRes
		}
	}

	sub.Data = args.Data
	if err := dao.SubmissionDbDefault().UpdateData(sub, args.OperatorId, args.OperatorName, ctx.ClientIP()); err != nil {
		cbase.PanicMsg(err.Error())
	}

	return sub, &engine.ValidationResult{IsValid: true}
}

func (DynamicServer) DeleteService(formId, id, opId, opName, clientIp string) {
	if err := dao.SubmissionDbDefault().DeleteById(formId, id, opId, opName, clientIp); err != nil {
		cbase.PanicMsg(err.Error())
	}
}
