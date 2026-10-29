package router

import (
	"hr-dynamic-form-engine/src/router/controller"

	"github.com/gin-gonic/gin"
)

func auditRouter(group *gin.RouterGroup) {
	auditCtrl := controller.AuditController{}
	auditGroup := group.Group("audit-logs")
	{
		auditGroup.GET("", auditCtrl.Query)
	}
}
