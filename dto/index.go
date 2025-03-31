package dto

import (
	"github.com/gin-gonic/gin"
)

type ResponseCode int

const (
	Fail ResponseCode = iota
	TokenExpired
	TokenError
	NotFound = 10
	Success  = 100
)

// 约定code
// 0: 失败
// 1: token过期，前端用存储的账号密码重新登录/跳回登录页
// 2: token错误，前端清除存储的账号密码，跳回登录页
// 100: 成功

type Response struct {
	Message string `json:"message"`
	Code	ResponseCode `json:"code"`
	Data    interface{} `json:"data"`
}

type ResponseOptions func(*Response)

func WithMessage(message string) ResponseOptions {
	return func(r *Response) {
		r.Message = message
	}
}

func WithCode(code ResponseCode) ResponseOptions {
	return func(r *Response) {
		r.Code = code
	}
}

func WithData(data any) ResponseOptions {
	return func(r *Response) {
		r.Data = data
	}
}

func SuccessResponse(c *gin.Context, opts ...ResponseOptions) {
	response := Response{
		Message: "success",
		Code: Success,
		Data: nil,
	}
	for _, opt := range opts {
		opt(&response)
	}
	c.JSON(200, response)
}

func ErrorResponse(c *gin.Context, opts ...ResponseOptions) {
	response := Response{
		Message: "error",
		Code: Fail,
		Data: nil,
	}
	for _, opt := range opts {
		opt(&response)
	}
	c.JSON(200, response)
}