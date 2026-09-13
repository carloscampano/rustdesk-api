package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestWebClientPublicPath(t *testing.T) {
	if !isWebClientPublicPath("/webclient/manifest.json") {
		t.Fatal("manifest.json should be public")
	}
	if !isWebClientPublicPath("/webclient/icons/Icon-192.png") {
		t.Fatal("icons should be public")
	}
	if isWebClientPublicPath("/webclient/") || isWebClientPublicPath("/webclient/main.dart.js") {
		t.Fatal("app files must stay behind login")
	}
}

func TestRequestToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/logout", nil)
	c.Request.Header.Set("api-token", "abc")
	if got := RequestToken(c); got != "abc" {
		t.Fatalf("api-token: got %q", got)
	}

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/logout", nil)
	c.Request.Header.Set("Authorization", "Bearer xyz")
	if got := RequestToken(c); got != "xyz" {
		t.Fatalf("bearer: got %q", got)
	}

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/logout", nil)
	c.Request.AddCookie(&http.Cookie{Name: WebClientCookie, Value: "cook"})
	if got := RequestToken(c); got != "cook" {
		t.Fatalf("cookie: got %q", got)
	}
}

func TestSetPortalCookieHttpOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/login", nil)
	c.Request.Header.Set("X-Forwarded-Proto", "https")
	SetPortalCookie(c, "secret-token")
	h := w.Header().Get("Set-Cookie")
	if h == "" {
		t.Fatal("missing Set-Cookie")
	}
	if !strings.Contains(strings.ToLower(h), "httponly") {
		t.Fatalf("cookie must be HttpOnly: %s", h)
	}
	if !strings.Contains(h, "rd_portal=secret-token") {
		t.Fatalf("cookie value: %s", h)
	}
	if !strings.Contains(strings.ToLower(h), "secure") {
		t.Fatalf("https cookie must be Secure: %s", h)
	}
}

func TestClearPortalCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/logout", nil)
	ClearPortalCookie(c)
	h := w.Header().Get("Set-Cookie")
	if h == "" {
		t.Fatal("missing Set-Cookie")
	}
	if !strings.Contains(h, "Max-Age=0") && !strings.Contains(strings.ToLower(h), "max-age=-1") {
		t.Fatalf("cookie should expire: %s", h)
	}
	if !strings.Contains(strings.ToLower(h), "httponly") {
		t.Fatalf("clear cookie must stay HttpOnly: %s", h)
	}
}
