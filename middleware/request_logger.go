package middleware

import (
	"bytes"
	"io"
	"web-api/util"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func RequestLogger(logMsg string) gin.HandlerFunc {
	return func(c *gin.Context) {
		buf, _ := io.ReadAll(c.Request.Body)
		rdr1 := io.NopCloser(bytes.NewBuffer(buf))
		rdr2 := io.NopCloser(bytes.NewBuffer(buf)) //We have to create a new Buffer, because rdr1 will be read.

		requestBody := readBody(rdr1) // Print request body
		c.Request.Body = rdr2
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
			"client_ip":     clientIP,
			"form":          form.Encode(),
			"request_body":  requestBody,
		}).Info(logMsg)
		c.Next()
	}
}