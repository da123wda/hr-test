package gin_utils

import (
	"encoding/json"
	"hr-dynamic-form-engine/src/cbase"
	"hr-dynamic-form-engine/src/log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Bind 绑定结构体请求参数并在失败时触发异常
func Bind[T any](ctx *gin.Context) T {
	t := new(T)
	if err := ctx.ShouldBind(t); err != nil {
		cbase.PanicMsg("参数格式错误: " + err.Error())
	}
	log.Zap.Infof("req args %+v, req path %v", *t, ctx.Request.URL.Path)
	return *t
}

// BindJSON 绑定 JSON 请求体
func BindJSON[T any](ctx *gin.Context) T {
	t := new(T)
	if err := ctx.ShouldBindJSON(t); err != nil {
		cbase.PanicMsg("JSON 请求参数错误: " + err.Error())
	}
	log.Zap.Infof("req json args %+v, req path %v", *t, ctx.Request.URL.Path)
	return *t
}

var okJson = func() []byte {
	b, _ := json.Marshal(cbase.ResVO[cbase.Unit]{Msg: "ok", Code: cbase.OkCode})
	return b
}()

func Ok_() (int, []byte) {
	return http.StatusOK, okJson
}

func Ok[T any](data T) (int, cbase.ResVO[T]) {
	return http.StatusOK, cbase.ResVO[T]{Msg: "ok", Code: cbase.OkCode, Data: data}
}

func ValidationError(msg string, errors any) (int, cbase.ResVO[any]) {
	return http.StatusOK, cbase.ResVO[any]{
		Code:   cbase.ExceptCode,
		Msg:    msg,
		Errors: errors,
	}
}
