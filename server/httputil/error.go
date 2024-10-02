package httputil

import "github.com/gin-gonic/gin"

func NewError(ctx *gin.Context, status int, message string) {
	er := HTTPError{
		Code:    status,
		Message: message,
	}
	ctx.JSON(status, er)
}

type HTTPError struct {
	Code    int    `json:"code" example:"400"`
	Message string `json:"message" example:"Bad Request"`
}

type HTTPError200 struct {
	Code    int    `json:"code" example:"200"`
	Message string `json:"message" example:"Success"`
}

type HTTPError400 struct {
	Code    int    `json:"code" example:"400"`
	Message string `json:"message" example:"Bad Request"`
}

type HTTPError401 struct {
	Code    int    `json:"code" example:"401"`
	Message string `json:"message" example:"Unauthorized"`
}

type HTTPError403 struct {
	Code    int    `json:"code" example:"403"`
	Message string `json:"message" example:"Forbidden"`
}

type HTTPError404 struct {
	Code    int    `json:"code" example:"404"`
	Message string `json:"message" example:"Not Found"`
}

type HTTPError405 struct {
	Code    int    `json:"code" example:"405"`
	Message string `json:"message" example:"Method Not Allowed"`
}

type HTTPError406 struct {
	Code    int    `json:"code" example:"406"`
	Message string `json:"message" example:"Not Acceptable"`
}

type HTTPError413 struct {
	Code    int    `json:"code" example:"413"`
	Message string `json:"message" example:"Request Entity Too Large"`
}

type HTTPError500 struct {
	Code    int    `json:"code" example:"500"`
	Message string `json:"message" example:"Internal Server Error"`
}
