package middleware

import (
	cache "github.com/chenyahui/gin-cache"
	"github.com/gin-gonic/gin"
	ca "web-api/cache"
	"time"
)

func Cache(duration time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		var prefix string
		if v, exists := c.Get("_tz_abbr"); exists {
			prefix += "tz" + v.(string)
		}
		if v, exists := c.Get("_app_type"); exists {
			prefix += v.(string)
		}
		prefix += c.MustGet("_language").(string)
		cache.CacheByRequestURI(ca.RedisStore, duration, cache.WithPrefixKey(prefix), cache.IgnoreQueryOrder())(c)
		c.Next()
	}
}