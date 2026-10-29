package controller

import (
	"fmt"
	"hr-dynamic-form-engine/src/cbase"
	. "hr-dynamic-form-engine/src/gin_utils"
	"hr-dynamic-form-engine/src/router/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type FileController struct{}

func (FileController) Upload(ctx *gin.Context) {
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		cbase.PanicMsg("请选择上传文件: " + err.Error())
	}
	formId := ctx.PostForm("formId")
	opId, _ := getOperatorInfo(ctx)

	resp := service.FileServer{}.UploadService(fileHeader, formId, opId)
	ctx.JSON(Ok(resp))
}

func (FileController) Download(ctx *gin.Context) {
	fileId := ctx.Param("fileId")
	if fileId == "" {
		cbase.PanicMsg("fileId 不能为空")
	}

	filePath, meta, err := service.FileServer{}.GetFileForDownload(fileId)
	if err != nil {
		ctx.String(http.StatusNotFound, err.Error())
		return
	}

	ctx.Header("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, meta.OriginalName))
	ctx.Header("Content-Type", meta.MimeType)
	ctx.File(filePath)
}
