package middleware

import (
	"bytes"
	"io"
	"web-api/conf/consts"
	"web-api/util"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func ErrorLogStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw
		buf, _ := io.ReadAll(c.Request.Body)
		rdr1 := io.NopCloser(bytes.NewBuffer(buf))
		rdr2 := io.NopCloser(bytes.NewBuffer(buf)) //We have to create a new Buffer, because rdr1 will be read.

		requestBody := readBody(rdr1) // Print request body
		c.Request.Body = rdr2

		c.Next()
		statusCode := c.Writer.Status()
		ginErr, exists := c.Get(consts.GinErrorKey)
		if statusCode >= 400 || exists {
			form := c.Request.Form
			// Request method
			reqMethod := c.Request.Method

			// Request route
			reqUri := c.Request.RequestURI

			// Request IP
			clientIP := c.ClientIP()
			util.GetLoggerEntry(c).WithFields(logrus.Fields{
				"method":        reqMethod,
				"uri":           reqUri,
				"status":        statusCode,
				"client_ip":     clientIP,
				"form":          form.Encode(),
				"response_body": blw.body.String(),
				"request_body":  requestBody,
				"error":         ginErr,
			}).Errorf("Error in response")
		}
	}
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func readBody(reader io.Reader) string {
	buf := new(bytes.Buffer)
	buf.ReadFrom(reader)

	s := buf.String()
	return s
}
