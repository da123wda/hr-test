package controller

import (
	"encoding/json"
	"hr-dynamic-form-engine/src/cbase"
	"hr-dynamic-form-engine/src/dto"
	. "hr-dynamic-form-engine/src/gin_utils"
	"hr-dynamic-form-engine/src/router/service"
	"strings"

	"github.com/gin-gonic/gin"
)

type DynamicController struct{}

func (DynamicController) Submit(ctx *gin.Context) {
	formId := ctx.Param("formId")
	if formId == "" {
		cbase.PanicMsg("formId 不能为空")
	}

	var rawBody map[string]interface{}
	if err := ctx.ShouldBindJSON(&rawBody); err != nil {
		cbase.PanicMsg("JSON 请求参数错误: " + err.Error())
	}

	req := dto.SubmitDataReq{
		SubmitterId:   ctx.GetString("UserId"),
		SubmitterName: ctx.GetString("UserName"),
		Data:          rawBody,
	}

	if innerData, ok := rawBody["data"].(map[string]interface{}); ok && len(rawBody) <= 3 {
		req.Data = innerData
		if subId, ok := rawBody["submitterId"].(string); ok {
			req.SubmitterId = subId
		}
		if subName, ok := rawBody["submitterName"].(string); ok {
			req.SubmitterName = subName
		}
	}

	resp, valRes := service.DynamicServer{}.SubmitService(ctx, formId, &req)
	if valRes != nil && !valRes.IsValid {
		ctx.JSON(ValidationError("表单数据校验失败，请检查各字段输入", valRes.Errors))
		return
	}

	ctx.JSON(Ok(resp))
}

func (DynamicController) Query(ctx *gin.Context) {
	formId := ctx.Param("formId")
	if formId == "" {
		cbase.PanicMsg("formId 不能为空")
	}

	page := getQueryInt(ctx, "page", 1)
	pageSize := getQueryInt(ctx, "pageSize", 10)
	schemaVer := getQueryInt(ctx, "schemaVersion", 0)
	status := ctx.Query("status")
	keyword := ctx.Query("keyword")

	dynamicFilters := make(map[string]interface{})
	filtersParam := ctx.Query("filters")
	if filtersParam != "" {
		_ = json.Unmarshal([]byte(filtersParam), &dynamicFilters)
	}

	for qKey, qValues := range ctx.Request.URL.Query() {
		if len(qValues) == 0 {
			continue
		}
		val := qValues[0]

		if strings.HasPrefix(qKey, "filter[") && strings.HasSuffix(qKey, "]") {
			fieldKey := qKey[7 : len(qKey)-1]
			if strings.Contains(val, ":") {
				parts := strings.SplitN(val, ":", 2)
				dynamicFilters[fieldKey] = map[string]interface{}{parts[0]: parts[1]}
			} else {
				dynamicFilters[fieldKey] = val
			}
		}

		for _, suffix := range []string{"_gt", "_gte", "_lt", "_lte", "_ne"} {
			if strings.HasSuffix(qKey, suffix) {
				fieldKey := strings.TrimSuffix(qKey, suffix)
				op := strings.TrimPrefix(suffix, "_")
				dynamicFilters[fieldKey] = map[string]interface{}{op: val}
			}
		}
	}

	filter := &dto.QuerySubmissionFilter{
		Page:          page,
		PageSize:      pageSize,
		SchemaVersion: schemaVer,
		Status:        status,
		Keyword:       keyword,
		Filters:       dynamicFilters,
	}

	res := service.DynamicServer{}.QueryService(ctx, formId, filter)
	ctx.JSON(Ok(res))
}

func (DynamicController) Get(ctx *gin.Context) {
	formId := ctx.Param("formId")
	id := ctx.Param("id")
	sub := service.DynamicServer{}.GetDetailService(formId, id)
	ctx.JSON(Ok(sub))
}

func (DynamicController) Update(ctx *gin.Context) {
	formId := ctx.Param("formId")
	id := ctx.Param("id")

	var rawBody map[string]interface{}
	if err := ctx.ShouldBindJSON(&rawBody); err != nil {
		cbase.PanicMsg("参数格式错误: " + err.Error())
	}

	opId, opName := getOperatorInfo(ctx)
	req := dto.UpdateDataReq{
		OperatorId:   opId,
		OperatorName: opName,
		Data:          rawBody,
	}
	if inner, ok := rawBody["data"].(map[string]interface{}); ok {
		req.Data = inner
	}

	updated, valRes := service.DynamicServer{}.UpdateService(ctx, formId, id, &req)
	if valRes != nil && !valRes.IsValid {
		ctx.JSON(ValidationError("更新数据校验未通过", valRes.Errors))
		return
	}

	ctx.JSON(Ok(updated))
}

func (DynamicController) Delete(ctx *gin.Context) {
	formId := ctx.Param("formId")
	id := ctx.Param("id")
	opId, opName := getOperatorInfo(ctx)

	service.DynamicServer{}.DeleteService(formId, id, opId, opName, ctx.ClientIP())
	ctx.JSON(Ok_())
}
