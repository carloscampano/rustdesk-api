package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lejianwen/rustdesk-api/v2/http/response"
)

const (
	csrfTokenHeader = "X-CSRF-Token"
	csrfCookieName  = "csrf_token"
	csrfTokenLength = 32
)

// generateCSRFToken creates a cryptographically secure random token
func generateCSRFToken() (string, error) {
	b := make([]byte, csrfTokenLength)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// CSRFProtection provides CSRF protection for state-changing requests.
// It validates that the X-CSRF-Token header matches the csrf_token cookie.
// GET/HEAD/OPTIONS requests get a new token set in a cookie.
// POST/PUT/DELETE/PATCH requests must include a valid token.
func CSRFProtection() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip CSRF for safe methods - just ensure they have a token
		if c.Request.Method == http.MethodGet ||
			c.Request.Method == http.MethodHead ||
			c.Request.Method == http.MethodOptions {
			// Generate new token if not present
			if _, err := c.Cookie(csrfCookieName); err != nil {
				token, err := generateCSRFToken()
				if err != nil {
					response.Fail(c, http.StatusInternalServerError, "Failed to generate CSRF token")
					c.Abort()
					return
				}
				// Set cookie accessible to JavaScript (not HttpOnly)
				c.SetCookie(csrfCookieName, token, 3600, "/", "", false, false)
				c.Header(csrfTokenHeader, token)
			}
			c.Next()
			return
		}

		// For state-changing methods, validate CSRF token
		cookieToken, err := c.Cookie(csrfCookieName)
		if err != nil || cookieToken == "" {
			response.Fail(c, http.StatusForbidden, "CSRF token missing")
			c.Abort()
			return
		}

		headerToken := c.GetHeader(csrfTokenHeader)
		if headerToken == "" {
			response.Fail(c, http.StatusForbidden, "CSRF token header missing")
			c.Abort()
			return
		}

		// Constant-time comparison to prevent timing attacks
		if !secureCompare(cookieToken, headerToken) {
			response.Fail(c, http.StatusForbidden, "CSRF token mismatch")
			c.Abort()
			return
		}

		c.Next()
	}
}

// secureCompare performs a constant-time comparison of two strings
func secureCompare(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}
	return result == 0
}

// CSRFExempt wraps a handler to skip CSRF validation for specific paths.
// Use this for API endpoints that use JWT authentication instead of cookies.
func CSRFExempt(exemptPaths ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, path := range exemptPaths {
			if strings.HasPrefix(c.Request.URL.Path, path) {
				c.Next()
				return
			}
		}
		CSRFProtection()(c)
	}
}
