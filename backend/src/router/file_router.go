package router

import (
	"hr-dynamic-form-engine/src/router/controller"

	"github.com/gin-gonic/gin"
)

func fileRouter(group *gin.RouterGroup) {
	fileCtrl := controller.FileController{}
	fileGroup := group.Group("files")
	{
		fileGroup.POST("upload", fileCtrl.Upload)
		fileGroup.GET("download/:fileId", fileCtrl.Download)
	}
}
