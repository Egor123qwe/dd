package resp_processor

import (
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-yaml"
	"net/http"
)

func YamlRespSender(c *gin.Context, data any, messages []string, err error) {
	if err == nil {
		resp, err := yaml.Marshal(data)
		if err != nil {
			JsonErrRespSender(c, messages, err)
			return
		}

		c.Data(http.StatusOK, "application/x-yaml", resp)
		return
	}

	JsonErrRespSender(c, messages, err)
}
