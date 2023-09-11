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
	logrus.SetOutput(os.Stdout)

	return func(c *gin.Context) {
		u := c.Request.Header.Get(consts.CorrelationHeader)
		if u == "" {
			u = uuid.NewString()
		}
		logrus.SetLevel(logrus.DebugLevel)
		contextLogger := logrus.WithField("correlation_id", u)
		c.Set(consts.LogKey, contextLogger)
		c.Set(consts.CorrelationKey, u)
		c.Header(consts.CorrelationHeader, u)
		c.Next()
	}
}
