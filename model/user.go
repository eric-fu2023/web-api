package model

import (
	models "blgit.rfdev.tech/taya/ploutos-object"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"os"
	"strconv"
	"time"
)

// User 用户模型
type User struct {
	models.UserC
	UserSum *UserSum `gorm:"foreignKey:UserId;references:ID"`
}

const (
	// PassWordCost 密码加密难度
	PassWordCost = 4
	// Active 激活用户
	Active string = "active"
	// Inactive 未激活用户
	Inactive string = "inactive"
	// Suspend 被封禁用户
	Suspend string = "suspend"
)

func (user *User) GenToken() (tokenString string, err error) {
	days, _ := strconv.Atoi(os.Getenv("TOKEN_EXPIRED_DAYS"))
	now := time.Now()
	exp := now.AddDate(0, 0, days)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":      user.ID,
		"country_code": user.CountryCode,
		"mobile":       user.Mobile,
		"iat":          now.Unix(),
		"exp":          exp.Unix(),
	})

	tokenString, err = token.SignedString([]byte(os.Getenv("SELF_JWT_HMAV_SECRET")))
	return
}

func (user *User) GetRedisSessionKey() string {
	return fmt.Sprintf(`session:%d`, user.ID)
}
