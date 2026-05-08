package server

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
)

func applyHeaders(c *gin.Context, raw []byte) {
	if len(raw) == 0 {
		return
	}

	headers := map[string]string{}
	if err := json.Unmarshal(raw, &headers); err != nil {
		return
	}

	for k, v := range headers {
		c.Header(k, v)
	}
}