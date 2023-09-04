package serializer

import (
	"github.com/gin-gonic/gin"
	"os"
	"web-api/model"
)

type User struct {
	ID          int64  `json:"id"`
	CountryCode string `json:"country_code,omitempty"`
	Mobile      string `json:"mobile,omitempty"`
	Username    string `json:"username,omitempty"`
	Email       string `json:"email,omitempty"`
	SmsOtp      string `json:"sms_otp,omitempty"`
	EmailOtp    string `json:"email_otp,omitempty"`
}

func BuildUser(user model.User) User {
	u := User{
		ID:          user.ID,
		CountryCode: user.CountryCode,
		Mobile:      user.Mobile,
		Username:    user.Username,
		Email:       user.Email,
	}
	if os.Getenv("ENV") == "staging" || os.Getenv("ENV") == "local" { // for development convenience
		//u.SmsOtp = user.SmsOtp
		//u.EmailOtp = user.EmailOtp
	}

	return u
}

type UserInfo struct {
	ID             int64    `json:"id"`
	CountryCode    string   `json:"country_code,omitempty"`
	Mobile         string   `json:"mobile,omitempty"`
	Username       string   `json:"username,omitempty"`
	Email          string   `json:"email,omitempty"`
	Avatar         string   `json:"avatar"`
	Bio            string   `json:"bio"`
	CurrencyId     int64    `json:"currency_id"`
	FollowingCount int64    `json:"following_count"`
	SetupRequired  bool     `json:"setup_required"`
	UserSum        *UserSum `json:"sum,omitempty"`
}

func BuildUserInfo(c *gin.Context, user model.User) UserInfo {
	u := UserInfo{
		ID:             user.ID,
		CountryCode:    user.CountryCode,
		Mobile:         user.Mobile,
		Username:       user.Username,
		Email:          user.Email,
		Bio:            user.Bio,
		CurrencyId:     user.CurrencyId,
		FollowingCount: user.FollowingCount,
	}
	if user.Avatar != "" {
		u.Avatar = Url(user.Avatar)
	}
	if user.Username == "" {
		u.SetupRequired = true
	}
	if user.UserSum != nil {
		t := BuildUserSum(*user.UserSum)
		u.UserSum = &t
	}

	return u
}
