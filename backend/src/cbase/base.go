package cbase

import "fmt"

const (
	OkCode     = 1
	ExceptCode = 10004
	ErrorCode  = 10005
)

type Unit struct{}

type CodeErr struct {
	Code int32
	Msg  string
}

func CodeErr_(code int32, msg string) CodeErr {
	return CodeErr{Code: code, Msg: msg}
}

func PanicMsg(msg string) {
	panic(CodeErr_(ExceptCode, msg))
}

func PanicCodeMsg(code int32, msg string) {
	panic(CodeErr_(code, msg))
}

func (e CodeErr) Error() string {
	return fmt.Sprintf("%v ~ %v", e.Code, e.Msg)
}

type ResVO[T any] struct {
	Msg    string `json:"msg"`
	Code   int32  `json:"code"`
	Data   T      `json:"data"`
	Errors any    `json:"errors,omitempty"`
}
