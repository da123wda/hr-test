package controller

import (
	"fmt"
	"hr-dynamic-form-engine/src/cbase"
	"hr-dynamic-form-engine/src/dto"
	. "hr-dynamic-form-engine/src/gin_utils"
	"hr-dynamic-form-engine/src/router/service"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type SchemaController struct{}

func (SchemaController) Get(ctx *gin.Context) {
	formId := ctx.Param("formId")
	if formId == "" {
		cbase.PanicMsg("formId 不能为空")
	}

	mode := ctx.Query("mode")
	var resp *dto.SchemaResp
	if mode == "draft" {
		resp = service.SchemaServer{}.GetDraftService(formId)
	} else {
		resp = service.SchemaServer{}.GetService(formId)
	}

	ctx.Header("ETag", resp.Etag)
	ctx.JSON(Ok(resp))
}

func (SchemaController) GetDraft(ctx *gin.Context) {
	formId := ctx.Param("formId")
	if formId == "" {
		cbase.PanicMsg("formId 不能为空")
	}
	resp := service.SchemaServer{}.GetDraftService(formId)
	ctx.Header("ETag", resp.Etag)
	ctx.JSON(Ok(resp))
}

func (SchemaController) Save(ctx *gin.Context) {
	args := BindJSON[dto.SaveSchemaReq](ctx)

	ifMatch := ctx.GetHeader("If-Match")
	if ifMatch != "" && args.ExpectedVersion == 0 {
		clean := strings.Trim(ifMatch, `W/"`)
		parts := strings.Split(clean, "-v")
		if len(parts) == 2 {
			if v, err := strconv.Atoi(parts[1]); err == nil {
				args.ExpectedVersion = v
			}
		}
	}

	opId, _ := getOperatorInfo(ctx)
	schema := service.SchemaServer{OperatorId: opId}.SaveService(&args)
	ctx.Header("ETag", fmt.Sprintf(`W/"%s-v%d"`, schema.FormId, schema.Version))
	ctx.JSON(Ok(schema))
}

func (SchemaController) Publish(ctx *gin.Context) {
	formId := ctx.Param("formId")
	if formId == "" {
		cbase.PanicMsg("formId 不能为空")
	}
	var args dto.PublishSchemaReq
	_ = ctx.ShouldBindJSON(&args)

	opId, _ := getOperatorInfo(ctx)
	schema := service.SchemaServer{OperatorId: opId}.PublishService(formId, &args)
	ctx.Header("ETag", fmt.Sprintf(`W/"%s-v%d"`, schema.FormId, schema.Version))
	ctx.JSON(Ok(schema))
}

func (SchemaController) ListVersions(ctx *gin.Context) {
	formId := ctx.Param("formId")
	if formId == "" {
		cbase.PanicMsg("formId 不能为空")
	}
	list := service.SchemaServer{}.ListVersionsService(formId)
	ctx.JSON(Ok(list))
}

func (SchemaController) GetVersion(ctx *gin.Context) {
	formId := ctx.Param("formId")
	verStr := ctx.Param("version")
	ver, err := strconv.Atoi(verStr)
	if err != nil {
		cbase.PanicMsg("版本号格式错误")
	}
	snap := service.SchemaServer{}.GetVersionService(formId, ver)
	ctx.JSON(Ok(snap))
}

func (SchemaController) ListAll(ctx *gin.Context) {
	list := service.SchemaServer{}.ListAllService()
	ctx.JSON(Ok(list))
}
