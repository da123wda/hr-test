package router

import (
	"hr-dynamic-form-engine/src/router/controller"

	"github.com/gin-gonic/gin"
)

func schemaRouter(group *gin.RouterGroup) {
	schemaCtrl := controller.SchemaController{}
	schemaGroup := group.Group("schemas")
	{
		schemaGroup.GET("", schemaCtrl.ListAll)
		schemaGroup.POST("", schemaCtrl.Save)
		schemaGroup.GET(":formId", schemaCtrl.Get)
		schemaGroup.GET(":formId/draft", schemaCtrl.GetDraft)
		schemaGroup.POST(":formId/publish", schemaCtrl.Publish)
		schemaGroup.GET(":formId/versions", schemaCtrl.ListVersions)
		schemaGroup.GET(":formId/versions/:version", schemaCtrl.GetVersion)
	}
}
