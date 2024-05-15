package middleware

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"os"
	"strconv"
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

func AuthRequired(getUser bool, checkBrand bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		i18n := c.MustGet("i18n").(i18n.I18n)
		err := doAuth(c, getUser, checkBrand)
		if err != nil {
			c.JSON(401, serializer.Response{
				Code:  serializer.CodeCheckLogin,
				Msg:   i18n.T("operation_not_allowed"),
				Error: err.Error(),
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

func CheckAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		doAuth(c, true, true)
		c.Next()
	}
}

func doAuth(c *gin.Context, getUser bool, checkBrand bool) (err error) {
	const BEARER_SCHEMA = "Bearer"
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		err = errors.New("no Authorization header")
		return
	}
	tokenString := strings.TrimSpace(authHeader[len(BEARER_SCHEMA):])
	c.Set("_token_string", tokenString)
	token, err := jwt.ParseWithClaims(tokenString, &AuthClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("SELF_JWT_HMAV_SECRET")), nil
	})
	if err != nil {
		return
	}
	a := token.Claims.(*AuthClaims)
	sess := cache.RedisSessionClient.HGet(context.TODO(), a.GetRedisSessionKey(), "token")
	if sess.Val() != tokenString {
		err = errors.New("invalid token")
		return
	}
	go func() {
		if timeout, e := strconv.Atoi(os.Getenv("SESSION_TIMEOUT")); e == nil {
			cache.RedisSessionClient.Expire(context.TODO(), a.GetRedisSessionKey(), time.Duration(timeout)*time.Minute)
		}
	}()
	if getUser {
		var user model.User
		if err = model.DB.Where(`id`, a.UserId).First(&user).Error; err != nil {
			return
		}
		if checkBrand {
			brand := c.MustGet(`_brand`).(int)
			if user.BrandId != int64(brand) {
				err = errors.New("user brand mismatch")
				return
			}
		}
		c.Set("user", user)

		if user.Locale != c.MustGet("_locale").(string) {
			go model.DB.Model(&user).Update(`locale`, c.MustGet("_locale").(string))
		}
	}
	return
}
