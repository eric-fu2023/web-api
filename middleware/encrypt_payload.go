package middleware

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"os"
	"web-api/util"
)

type customWriter struct {
	gin.ResponseWriter
	buf *bytes.Buffer
}

func (cw customWriter) Write(b []byte) (int, error) {
	return cw.buf.Write(b)
}

func EncryptPayload() gin.HandlerFunc {
	return func(c *gin.Context) {
		signature := c.GetHeader("Signature")
		timestamp := c.GetHeader("Timestamp")
		if signature == os.Getenv("SUPER_SIGNATURE") {
			return
		}
		if signature == "" || timestamp == "" {
			c.Abort()
			return
		}
		cw := &customWriter{buf: &bytes.Buffer{}, ResponseWriter: c.Writer}
		c.Writer = cw
		c.Next()
		if cw.buf != nil {
			body, e := util.AesCFBModeEncrypt(cw.buf.Bytes())
			if e != nil {
				c.Abort()
				return
			}
			cw.buf = bytes.NewBuffer([]byte(body))
			//cw.WriteHeader(c.Writer.Status())
			cw.WriteString(body)
		}
	}
}
