package middleware

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/gin-gonic/gin"
	"os"
	"sort"
	"strconv"
	"time"
)

func CheckSignature() gin.HandlerFunc {
	return func(c *gin.Context) {
		signature := c.GetHeader("Signature")
		timestamp := c.GetHeader("Timestamp")
		if signature == os.Getenv("SUPER_SIGNATURE") {
			c.Next()
			return
		}
		if signature == "" || timestamp == "" {
			c.Abort()
			return
		}
		var keys []string
		for k, _ := range c.Request.URL.Query() {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var str string
		if c.Request.Method == "GET" {
			for _, k := range keys {
				for _, a := range c.Request.URL.Query()[k] {
					str += a
				}
			}
		}
		str += timestamp + os.Getenv("SIGNATURE_SALT")
		hash := md5.Sum([]byte(str))
		h := hex.EncodeToString(hash[:])
		if h != signature {
			if os.Getenv("ENV") == "local" || os.Getenv("ENV") == "staging" {
				c.JSON(200, map[string]string{
					"error": "Signature invalid",
					"str": str,
					"hash": h,
				})
			}
			c.Abort()
			return
		}
		now := time.Now().Unix()
		min := now - 30
		max := now + 30
		t, e := strconv.Atoi(timestamp)
		if e != nil {
			c.Abort()
			return
		}
		if int64(t) < min || int64(t) > max {
			if os.Getenv("ENV") == "local" || os.Getenv("ENV") == "staging" {
				c.JSON(200, map[string]string{
					"error": "Timestamp out of range",
				})
			}
			c.Abort()
			return
		}
		c.Next()
	}
}