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
	ID             int64      `json:"id"`
	CountryCode    string     `json:"country_code,omitempty"`
	Mobile         string     `json:"mobile,omitempty"`
	Username       string     `json:"username,omitempty"`
	Email          string     `json:"email,omitempty"`
	Avatar         string     `json:"avatar"`
	Bio            string     `json:"bio"`
	FollowingCount int64      `json:"following_count"`
	SetupRequired  bool       `json:"setup_required"`
}

func BuildUserInfo(c *gin.Context, user model.User) UserInfo {
	u := UserInfo{
		ID:             user.ID,
		CountryCode:    user.CountryCode,
		Mobile:         user.Mobile,
		Username:       user.Username,
		Email:          user.Email,
		Bio:            user.Bio,
		FollowingCount: user.FollowingCount,
	}
	if user.Avatar != "" {
		u.Avatar = Url(user.Avatar)
	}
	if user.Username == "" {
		u.SetupRequired = true
	}

	return u
}
