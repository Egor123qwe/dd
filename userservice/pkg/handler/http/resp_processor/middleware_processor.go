package resp_processor

import (
	"github.com/gin-gonic/gin"
)

func MiddlewareProcessor(c *gin.Context, data any, messages []string, err error) {
	if err == nil {
		c.Next()
		return
	}

	JsonErrRespSender(c, messages, err)
	c.Abort()
}
