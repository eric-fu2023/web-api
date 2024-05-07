package middleware

import (
	"fmt"
	cache "github.com/chenyahui/gin-cache"
	"github.com/gin-gonic/gin"
	"time"
	ca "web-api/cache"
)

func Cache(duration time.Duration, needAbCache bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var prefix string
		if v, exists := c.Get("_tz_abbr"); exists {
			prefix += "tz" + v.(string)
		}
		if v, exists := c.Get("_brand"); exists {
			if vv, ok := v.(int); ok {
				prefix += fmt.Sprintf(`%d`, vv)
			}
		}
		if v, exists := c.Get("_channel"); exists {
			prefix += fmt.Sprintf(`%s`, v)
		}
		//if v, exists := c.Get("_agent"); exists {
		//	if vv, ok := v.(int); ok {
		//		prefix += fmt.Sprintf(`_%d`, vv)
		//	}
		//}
		prefix += c.MustGet("_language").(string)

		v, exists := c.Get("_isA")
		if needAbCache && exists {
			if isA, ok := v.(bool); ok && isA {
				prefix += "_a"
			}
		}

		cache.CacheByRequestURI(ca.RedisStore, duration, cache.WithPrefixKey(prefix), cache.IgnoreQueryOrder())(c)
	}
}

func CacheForGuest(duration time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, loggedIn := c.Get("user")
		if loggedIn {
			return
		} else {
			Cache(duration, false)(c)
		}
	}
}
