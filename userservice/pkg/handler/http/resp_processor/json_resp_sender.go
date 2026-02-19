package resp_processor

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func JsonRespSender(c *gin.Context, data any, messages []string, err error) {
	if err != nil {
		JsonErrRespSender(c, messages, err)
		return
	}

	result := Response{
		Data:       data,
		Messages:   messages,
		Validation: make(map[string][]string),
	}

	c.JSON(http.StatusOK, result)
}
