package router

import (
	"hr-dynamic-form-engine/src/cfg"
	"hr-dynamic-form-engine/src/log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Init() {
	switch cfg.Cfg.Env {
	case "dev":
		gin.SetMode(gin.DebugMode)
	case "test":
		gin.SetMode(gin.TestMode)
	case "prod":
		gin.SetMode(gin.ReleaseMode)
	}

	e := gin.New()
	e.UseH2C = true
	_ = e.SetTrustedProxies([]string{"127.0.0.1"})

	e.Use(corsMiddleware())
	e.Use(log.ErrHandler)

	e.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	g := e.Group("/api/v1")
	{
		schemaRouter(g)
		dynamicRouter(g)
		fileRouter(g)
		auditRouter(g)
	}

	log.Zap.Infof("HTTP server starting at :%s", cfg.Cfg.Port)
	if err := e.Run(":" + cfg.Cfg.Port); err != nil {
		log.Zap.Fatalf("HTTP server run error: %v", err)
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
			c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization, If-Match, X-Operator-Id, X-Operator-Name")
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type, ETag")
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
