package service

import (
	"fmt"
	"hr-dynamic-form-engine/src/cbase"
	"hr-dynamic-form-engine/src/dao"
	"hr-dynamic-form-engine/src/dto"
	"hr-dynamic-form-engine/src/engine"
	"hr-dynamic-form-engine/src/model"
	"strings"

	"github.com/google/uuid"
)

type SchemaServer struct {
	OperatorId string
}

// GetService 获取对外发布的正式表单定义 (面向线上填表用户)
func (s SchemaServer) GetService(formId string) *dto.SchemaResp {
	schema, err := dao.FormSchemaDbDefault().GetByFormId(formId)
	if err != nil {
		cbase.PanicMsg(fmt.Sprintf("未找到表单 [%s] 的定义", formId))
	}

	if schema.PublishedVersion == 0 {
		cbase.PanicMsg(fmt.Sprintf("表单 [%s] 尚未发布，暂不可用", formId))
	}

	// 读取当前发布的正式版本快照，隔离 HR 未发布的草稿改动
	snap, err := dao.FormSchemaDbDefault().GetVersion(formId, schema.PublishedVersion)
	if err == nil && snap != nil {
		schema.Fields = snap.Fields
		schema.Title = snap.Title
		schema.Description = snap.Description
	}

	etag := fmt.Sprintf(`W/"%s-v%d"`, schema.FormId, schema.PublishedVersion)
	return &dto.SchemaResp{
		FormSchema: *schema,
		Etag:       etag,
	}
}

// GetDraftService 获取工作台编辑草稿 (面向 HR 设计器)
func (s SchemaServer) GetDraftService(formId string) *dto.SchemaResp {
	schema, err := dao.FormSchemaDbDefault().GetByFormId(formId)
	if err != nil {
		cbase.PanicMsg(fmt.Sprintf("未找到表单 [%s] 的定义", formId))
	}
	etag := fmt.Sprintf(`W/"%s-v%d"`, schema.FormId, schema.Version)
	return &dto.SchemaResp{
		FormSchema: *schema,
		Etag:       etag,
	}
}

func (s SchemaServer) SaveService(args *dto.SaveSchemaReq) *model.FormSchema {
	ruleEngine := engine.NewRuleEngine()
	if hasCycle, path := ruleEngine.DetectCycles(args.Fields); hasCycle {
		cbase.PanicMsg(fmt.Sprintf("表单规则检测到循环依赖闭环: %s", strings.Join(path, " -> ")))
	}

	existing, err := dao.FormSchemaDbDefault().GetByFormId(args.FormId)
	var schema *model.FormSchema
	if err != nil {
		schema = &model.FormSchema{
			Id:          "schema_" + uuid.New().String()[:8],
			FormId:      args.FormId,
			Title:       args.Title,
			Description: args.Description,
			Version:     1,
			Status:      model.SchemaStatusDraft,
			Fields:      args.Fields,
			Migrations:  args.Migrations,
			CreatedBy:   s.OperatorId,
		}
	} else {
		schema = existing
		schema.Title = args.Title
		schema.Description = args.Description
		schema.Fields = args.Fields
		schema.Migrations = args.Migrations
	}

	if err := dao.FormSchemaDbDefault().SaveSchema(schema, args.ExpectedVersion); err != nil {
		cbase.PanicMsg(err.Error())
	}

	_ = dao.AuditDbDefault().InsertOne(&model.AuditLog{
		Action:       "UPDATE_SCHEMA",
		TargetType:   "SCHEMA",
		TargetId:     schema.FormId,
		FormId:       schema.FormId,
		OperatorId:   s.OperatorId,
		OperatorName: s.OperatorId,
	})

	return schema
}

func (s SchemaServer) PublishService(formId string, args *dto.PublishSchemaReq) *model.FormSchema {
	schema, err := dao.FormSchemaDbDefault().PublishSchema(formId, args.ExpectedVersion, s.OperatorId)
	if err != nil {
		cbase.PanicMsg(err.Error())
	}

	_ = dao.AuditDbDefault().InsertOne(&model.AuditLog{
		Action:       "PUBLISH_SCHEMA",
		TargetType:   "SCHEMA",
		TargetId:     formId,
		FormId:       formId,
		OperatorId:   s.OperatorId,
		OperatorName: s.OperatorId,
	})

	return schema
}

func (SchemaServer) ListVersionsService(formId string) []model.FormSchemaVersion {
	list, err := dao.FormSchemaDbDefault().ListVersions(formId)
	if err != nil {
		cbase.PanicMsg(err.Error())
	}
	return list
}

func (SchemaServer) GetVersionService(formId string, version int) *model.FormSchemaVersion {
	snap, err := dao.FormSchemaDbDefault().GetVersion(formId, version)
	if err != nil {
		cbase.PanicMsg(err.Error())
	}
	return snap
}

func (SchemaServer) ListAllService() []model.FormSchema {
	list, err := dao.FormSchemaDbDefault().ListAll()
	if err != nil {
		cbase.PanicMsg(err.Error())
	}
	return list
}
