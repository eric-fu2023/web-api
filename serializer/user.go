package serializer

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/gin-gonic/gin"
	"os"
	"strings"
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
	Nickname       string   `json:"nickname"`
	Avatar         string   `json:"avatar"`
	Bio            string   `json:"bio"`
	CurrencyId     int64    `json:"currency_id"`
	Signature      string   `json:"signature,omitempty"`
	FollowingCount int64    `json:"following_count"`
	SetupRequired  bool     `json:"setup_required"`
	UserSum        *UserSum `json:"sum,omitempty"`
}

func BuildUserInfo(c *gin.Context, user model.User) UserInfo {
	signatureHash := md5.Sum([]byte(fmt.Sprintf("%d%s", user.ID, os.Getenv("USER_SIGNATURE_SALT"))))
	u := UserInfo{
		ID:             user.ID,
		CountryCode:    user.CountryCode,
		Mobile:         user.Mobile,
		Username:       user.Username,
		Email:          user.Email,
		Nickname:       user.Nickname,
		Avatar:         Url(user.Avatar),
		Bio:            user.Bio,
		CurrencyId:     user.CurrencyId,
		Signature:      hex.EncodeToString(signatureHash[:]),
		FollowingCount: user.FollowingCount,
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

type PersonalInfo struct {
	Nickname         string `json:"nickname"`
	Username         string `json:"username"`
	Email            string `json:"email"`
	CountryCode      string `json:"country_code"`
	Mobile           string `json:"mobile"`
	Avatar           string `json:"avatar"`
	FirstName        string `json:"first_name"`
	MiddleName       string `json:"middle_name"`
	LastName         string `json:"last_name"`
	PermanentAddress string `json:"permanent_address"`
	CurrentAddress   string `json:"current_address"`
	Birthday         string `json:"birthday"`
}

func BuildPersonalInfo(c *gin.Context, user model.User) PersonalInfo {
	u := PersonalInfo{
		Username:    getMaskedUsername(user.Username),
		Nickname:    user.Nickname,
		CountryCode: user.CountryCode,
		Mobile:      getMaskedMobile(user.Mobile),
		Email:       getMaskedEmail(user.Email),
		Avatar:      Url(user.Avatar),
	}
	if user.Kyc != nil {
		u.FirstName = user.Kyc.FirstName
		u.MiddleName = user.Kyc.MiddleName
		u.LastName = user.Kyc.LastName
		u.PermanentAddress = user.Kyc.PermanentAddress
		u.CurrentAddress = user.Kyc.CurrentAddress
		u.Birthday = user.Kyc.Birthday
	}

	return u
}

func getMaskedUsername(original string) (new string) {
	l := len(original)
	if l == 0 {
		return
	}
	q := l / 3
	r := l % 3
	if q >= 1 {
		ast := ""
		for i := 0; i < q+r; i++ {
			ast += "*"
		}
		new = original[:q] + ast + original[l-q:l]
	} else {
		new = original[:1] + "*"
	}
	return
}

func getMaskedEmail(original string) (new string) {
	if len(original) == 0 {
		return
	}
	l := strings.Index(original, "@")
	if l == -1 || l == 0 {
		new = original
		return
	}
	q := l / 2
	r := l % 2
	if q >= 1 {
		ast := ""
		for i := 0; i < q+r; i++ {
			ast += "*"
		}
		new = original[:q] + ast
	} else {
		new = original[:1] + "*"
	}
	new += original[l:]
	return
}

func getMaskedMobile(original string) (new string) {
	l := len(original)
	if l == 0 {
		return
	}
	q := l / 2
	r := l % 2
	if q >= 1 {
		ast := ""
		for i := 0; i < q+r; i++ {
			ast += "*"
		}
		new = original[:q] + ast
	} else {
		new = original[:1] + "*"
	}
	return
}
