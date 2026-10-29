package log

import (
	"hr-dynamic-form-engine/src/cbase"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var Zap *zap.SugaredLogger

func Init() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	Zap = logger.Sugar()
	Zap.Infoln("logger init success")
}

// ErrHandler 捕获业务异常并转换为统一错误格式
func ErrHandler(ctx *gin.Context) {
	defer func() {
		switch e := recover().(type) {
		case cbase.CodeErr:
			ctx.JSON(http.StatusOK, cbase.ResVO[any]{Code: e.Code, Msg: e.Msg})
			Zap.Errorf("业务异常状态码: %v, msg: %v", e.Code, e.Msg)
		case error:
			Zap.Errorf("Server error: %v", e.Error())
			ctx.JSON(http.StatusOK, cbase.ResVO[any]{Code: cbase.ErrorCode, Msg: e.Error()})
		case nil:
			break
		default:
			Zap.Errorf("Unknown error: %#v", e)
			ctx.JSON(http.StatusOK, cbase.ResVO[any]{Code: cbase.ErrorCode, Msg: "Unknown error"})
		}
	}()
	ctx.Next()
}
