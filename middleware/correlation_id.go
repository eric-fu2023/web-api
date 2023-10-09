// Package correlationid adds a request correlation UUID to the request context,
// and includes an optional RequestLogger middleware including the UUID in all
// request log statements.
package middleware

import (
	"os"
	"web-api/conf/consts"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// SetRequestUUID will search for a correlation header and set a request-level
// correlation ID into the net.Context. If no header is found, a new UUID will
// be generated.
func CorrelationID() gin.HandlerFunc {
	return func(c *gin.Context) {
		u := c.Request.Header.Get(consts.CorrelationHeader)
		if u == "" {
			u = uuid.NewString()
		}
		logger := logrus.New()
		logger.SetOutput(os.Stdout)
		logger.SetLevel(logrus.DebugLevel)
		contextLogger := logger.WithField("correlation_id", u)
		// contextLogger.Logger.SetLevel(logrus.DebugLevel)
		c.Set(consts.LogKey, contextLogger)
		c.Set(consts.CorrelationKey, u)
		c.Header(consts.CorrelationHeader, u)
		c.Next()
	}
}
