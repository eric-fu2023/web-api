package middleware

import (
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Cors 跨域配置
func Cors() gin.HandlerFunc {
	config := cors.DefaultConfig()
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Cookie", "Device-Info", "*"}
	if gin.Mode() != gin.ReleaseMode {
		config.AllowOriginFunc = func(origin string) bool {
			return true
		}
	}
	if os.Getenv("CORS_DOMAINS") != "" {
		config.AllowOrigins = strings.Split(os.Getenv("CORS_DOMAINS"), ",")
	}
	config.AllowCredentials = true
	return cors.New(config)
}
