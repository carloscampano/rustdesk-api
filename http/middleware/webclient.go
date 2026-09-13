package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lejianwen/rustdesk-api/v2/global"
	"github.com/lejianwen/rustdesk-api/v2/service"
)

const WebClientCookie = "rd_portal"
const webClientLogin = "/_admin/#/login?redirect=/webclient/"

func isWebClientPublicPath(path string) bool {
	p := strings.ToLower(path)
	if strings.HasSuffix(p, "/manifest.json") {
		return true
	}
	if strings.HasSuffix(p, "/favicon.svg") || strings.HasSuffix(p, "/favicon.ico") {
		return true
	}
	if strings.Contains(p, "/icons/") {
		return true
	}
	return false
}

func WebClientPortalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if isWebClientPublicPath(c.Request.URL.Path) {
			c.Next()
			return
		}
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

func RequestToken(c *gin.Context) string {
	if t := strings.TrimSpace(c.GetHeader("api-token")); t != "" {
		return t
	}
	auth := c.GetHeader("Authorization")
	if len(auth) > 7 && strings.EqualFold(auth[:7], "Bearer ") {
		return strings.TrimSpace(auth[7:])
	}
	if t, err := c.Cookie(WebClientCookie); err == nil {
		return strings.TrimSpace(t)
	}
	return ""
}

func PortalCookieMaxAge() int {
	d := time.Duration(0)
	if global.Config.App.TokenExpire > 0 {
		d = global.Config.App.TokenExpire
	}
	if d <= 0 {
		d = 14 * 24 * time.Hour
	}
	return int(d.Seconds())
}

func isHTTPS(c *gin.Context) bool {
	if c.Request != nil && c.Request.TLS != nil {
		return true
	}
	proto := c.GetHeader("X-Forwarded-Proto")
	if proto == "" && c.Request != nil {
		proto = c.Request.Header.Get("X-Forwarded-Proto")
	}
	return strings.EqualFold(proto, "https")
}

func SetPortalCookie(c *gin.Context, token string) {
	if token == "" {
		return
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     WebClientCookie,
		Value:    token,
		Path:     "/",
		MaxAge:   PortalCookieMaxAge(),
		HttpOnly: true,
		Secure:   isHTTPS(c),
		SameSite: http.SameSiteLaxMode,
	})
}

func ClearPortalCookie(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     WebClientCookie,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isHTTPS(c),
		SameSite: http.SameSiteLaxMode,
	})
}
