package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/lejianwen/rustdesk-api/v2/global"
	"github.com/lejianwen/rustdesk-api/v2/http/response"
	"net/http"
)

func Limiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		loginLimiter := global.LoginLimiter
		clientIp := c.ClientIP()
		banned, _ := loginLimiter.CheckSecurityStatus(clientIp)
		if banned {
			msg := response.TranslateMsg(c, "Banned")
			// Native RustDesk clients only look at HTTP status + "error".
			// HTTP 200 + {code:423} is parsed as a login payload and shown as
			// "Failed, bad response from server".
			c.JSON(http.StatusForbidden, gin.H{
				"code":    http.StatusForbidden,
				"message": msg,
				"error":   msg,
				"data":    nil,
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
