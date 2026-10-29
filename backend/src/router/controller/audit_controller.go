package controller

import (
	. "hr-dynamic-form-engine/src/gin_utils"
	"hr-dynamic-form-engine/src/router/service"

	"github.com/gin-gonic/gin"
)

type AuditController struct{}

func (AuditController) Query(ctx *gin.Context) {
	targetType := ctx.Query("targetType")
	targetId := ctx.Query("targetId")
	formId := ctx.Query("formId")
	page := getQueryInt(ctx, "page", 1)
	pageSize := getQueryInt(ctx, "pageSize", 15)

	res := service.AuditServer{}.QueryService(targetType, targetId, formId, page, pageSize)
	ctx.JSON(Ok(res))
}
