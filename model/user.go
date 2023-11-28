package model

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ploutos.User
	KycCheckRequired bool             `gorm:"-"`
	UserSum          *ploutos.UserSum `gorm:"foreignKey:UserId;references:ID"`
	Kyc              *Kyc             `gorm:"foreignKey:UserId;references:ID"`
}

const (
	PassWordCost        = 4
	Active       string = "active"
	Inactive     string = "inactive"
	Suspend      string = "suspend"
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

func (user *User) VerifySecondaryPassword(secondaryPassword string) (err error) {
	if secondaryPassword == "" {
		return errors.New("empty password")
	}
	return bcrypt.CompareHashAndPassword([]byte(user.SecondaryPassword), []byte(secondaryPassword))
}
