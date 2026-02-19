package response

import "github.com/gin-gonic/gin"

func NewError(ctx *gin.Context, statusCode int, msg string) {
	resp := HTTPError{
		Code:    statusCode,
		Message: msg,
	}

	ctx.JSON(statusCode, resp)
}

type HTTPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
