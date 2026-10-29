package router

import (
	"hr-dynamic-form-engine/src/router/controller"

	"github.com/gin-gonic/gin"
)

func dynamicRouter(group *gin.RouterGroup) {
	dynamicCtrl := controller.DynamicController{}
	dynamicGroup := group.Group("dynamic-data")
	{
		dynamicGroup.POST(":formId", dynamicCtrl.Submit)
		dynamicGroup.GET(":formId", dynamicCtrl.Query)
		dynamicGroup.GET(":formId/:id", dynamicCtrl.Get)
		dynamicGroup.PUT(":formId/:id", dynamicCtrl.Update)
		dynamicGroup.DELETE(":formId/:id", dynamicCtrl.Delete)
	}
}
