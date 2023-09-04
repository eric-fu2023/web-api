package middleware

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"os"
	"strings"
	"time"
	"web-api/cache"
	"web-api/model"
	"web-api/serializer"
	"web-api/util/i18n"
)

type AuthClaims struct {
	CountryCode string `json:"country_code"`
	Mobile      string `json:"mobile"`
	UserId      int    `json:"user_id"`
	Nickname    string `json:"nickname"`
	IssuedAt    int64  `json:"iat"`
	ExpiresAt   int64  `json:"exp"`
}

func (a AuthClaims) Valid() (err error) {
	now := time.Now()
	if a.ExpiresAt == 0 || now.After(time.Unix(a.ExpiresAt, 0)) {
		err = errors.New("exp can't be empty or before now")
		return
	}
	return
}

func (a AuthClaims) GetRedisSessionKey() string {
	return fmt.Sprintf(`session:%d`, a.UserId)
}

// AuthRequired 需要登录
func AuthRequired(getUser bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		i18n := c.MustGet("i18n").(i18n.I18n)
		const BEARER_SCHEMA = "Bearer"
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, serializer.Response{
				Code:  serializer.CodeCheckLogin,
				Msg:   i18n.T("Token无效"),
				Error: "no Authorization header",
			})
			c.Abort()
			return
		}
		tokenString := strings.TrimSpace(authHeader[len(BEARER_SCHEMA):])
		token, err := jwt.ParseWithClaims(tokenString, &AuthClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("SELF_JWT_HMAV_SECRET")), nil
		})
		if err != nil {
			c.JSON(401, serializer.Response{
				Code:  serializer.CodeCheckLogin,
				Msg:   i18n.T("Token无效"),
				Error: err.Error(),
			})
			c.Abort()
			return
		}
		a := token.Claims.(*AuthClaims)

		otp := cache.RedisSessionClient.Get(context.TODO(), a.GetRedisSessionKey())
		if otp.Val() != tokenString {
			c.JSON(401, serializer.Response{
				Code: serializer.CodeCheckLogin,
				Msg:  i18n.T("Token无效"),
			})
			c.Abort()
			return
		}

		if getUser {
			var user model.User
			if err := model.DB.Where(`id`, a.UserId).First(&user).Error; err != nil {
				c.JSON(401, serializer.Response{
					Code:  serializer.CodeCheckLogin,
					Msg:   i18n.T("账号错误"),
					Error: err.Error(),
				})
				c.Abort()
				return
			}
			c.Set("user", user)
		}

		go func() {
			cache.RedisSessionClient.Expire(context.TODO(), a.GetRedisSessionKey(), 20 * time.Minute)
		}()
		c.Next()
	}
}
