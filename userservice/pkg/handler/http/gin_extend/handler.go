package gin_extend

import (
	"github.com/gin-gonic/gin"
)

type HandlerFunc func(c *gin.Context) (any, []string, error)

type RespProcessorFunc func(c *gin.Context, data any, messages []string, err error)

func NewHandler(respSender RespProcessorFunc) func(h HandlerFunc) gin.HandlerFunc {
	return func(fn HandlerFunc) gin.HandlerFunc {
		return func(c *gin.Context) {
			data, messages, err := fn(c)

			respSender(c, data, messages, err)
		}
	}
}
