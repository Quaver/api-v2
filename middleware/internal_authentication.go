package middleware

import (
	"crypto/subtle"
	"net/http"

	"github.com/Quaver/api2/config"
	"github.com/gin-gonic/gin"
)

// HasValidInternalSecret checks only the dedicated internal API credential.
func HasValidInternalSecret(c *gin.Context) bool {
	if config.Instance == nil || config.Instance.Server.InternalAPISecret == "" {
		return false
	}

	return subtle.ConstantTimeCompare(
		[]byte(c.GetHeader("X-Internal-Secret")),
		[]byte(config.Instance.Server.InternalAPISecret),
	) == 1
}

func RequireInternalSecret(c *gin.Context) {
	if !HasValidInternalSecret(c) {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "You must provide a valid X-Internal-Secret."})
		return
	}

	c.Next()
}
