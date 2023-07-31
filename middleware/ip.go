package middleware

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/exp/slices"
	"time"
	"web-api/model"
)

var ipWhitelist map[string]interface{}
var ipBlacklist map[string]interface{}

func Ip() gin.HandlerFunc {
	return func(c *gin.Context) {
		now := time.Now()
		getWhitelist := false
		getBlacklist := false
		if ipWhitelist == nil || ipWhitelist["expiry"] == nil {
			getWhitelist = true
		} else {
			if now.After(ipWhitelist["expiry"].(time.Time)) {
				getWhitelist = true
			}
		}
		if ipBlacklist == nil || ipBlacklist["expiry"] == nil {
			getBlacklist = true
		} else {
			if now.After(ipBlacklist["expiry"].(time.Time)) {
				getBlacklist = true
			}
		}

		if getWhitelist {
			var cc []map[string]interface{}
			var ccc []string
			model.DB.Table("ips").Where(`status = 1`).Where(`type = 1`).Select("name").Find(&cc)
			for _, a := range cc {
				ccc = append(ccc, a["ip"].(string))
			}
			ipWhitelist = map[string]interface{}{
				"expiry": now.Add(5 * time.Minute),
				"values": ccc,
			}
		}
		if getBlacklist {
			var cc []map[string]interface{}
			var ccc []string
			model.DB.Table("ips").Where(`status = 1`).Where(`type = 0`).Select("name").Find(&cc)
			for _, a := range cc {
				ccc = append(ccc, a["ip"].(string))
			}
			ipBlacklist = map[string]interface{}{
				"expiry": now.Add(5 * time.Minute),
				"values": ccc,
			}
		}

		if slices.Contains(ipWhitelist["values"].([]string), c.ClientIP()) { // set in the context for whitelisting later in the controllers
			c.Set("_white_ip", c.ClientIP())
			return
		}

		if slices.Contains(ipBlacklist["values"].([]string), c.ClientIP()) { // stop the request if the ip is blacklisted
			c.Abort()
			return
		}
		c.Next()
	}
}
