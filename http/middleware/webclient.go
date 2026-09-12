package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lejianwen/rustdesk-api/v2/service"
)

const WebClientCookie = "rd_portal"
const webClientLogin = "/_admin/#/login?redirect=/webclient/"

func WebClientPortalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, _ := c.Cookie(WebClientCookie)
		if token == "" {
			denyWebClient(c)
			return
		}
		user, _ := service.AllService.UserService.InfoByAccessToken(token)
		if user.Id == 0 || !service.AllService.UserService.CheckUserEnable(user) {
			denyWebClient(c)
			return
		}
		c.Set("curUser", user)
		c.Next()
	}
}

func denyWebClient(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	c.Redirect(http.StatusFound, webClientLogin)
	c.Abort()
}
