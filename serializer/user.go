package serializer

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
	"os"
	"time"
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
		Mobile:      string(user.Mobile),
		Username:    user.Username,
		Email:       string(user.Email),
	}
	if os.Getenv("ENV") == "staging" || os.Getenv("ENV") == "local" { // for development convenience
		//u.SmsOtp = user.SmsOtp
		//u.EmailOtp = user.EmailOtp
	}

	return u
}

type UserInfo struct {
	ID                                int64    `json:"id"`
	CountryCode                       string   `json:"country_code,omitempty"`
	Mobile                            string   `json:"mobile,omitempty"`
	Username                          string   `json:"username,omitempty"`
	Email                             string   `json:"email,omitempty"`
	Nickname                          string   `json:"nickname"`
	Avatar                            string   `json:"avatar"`
	Bio                               string   `json:"bio"`
	CurrencyId                        int64    `json:"currency_id"`
	Signature                         string   `json:"signature,omitempty"`
	FollowingCount                    int64    `json:"following_count"`
	SetupRequired                     bool     `json:"setup_required"`
	KycRequired                       bool     `json:"kyc_required"`
	HasSetSecondaryPwd                bool     `json:"has_set_secondary_pwd"`
	Brand                             int64    `json:"brand"`
	Agent                             int64    `json:"agent"`
	UserSum                           *UserSum `json:"sum,omitempty"`
	Kyc                               *Kyc     `json:"kyc,omitempty"`
	HasCompletedFirstAppLoginTutorial bool     `json:"has_completed_first_app_login_tutorial"`
	HasClaimedFirstAppLoginReward     bool     `json:"has_claimed_first_app_login_reward"`
	Birthday                          string   `json:"birthday"`
}

func BuildUserInfo(c *gin.Context, user model.User) UserInfo {
	u := UserInfo{
		ID:                 user.ID,
		CountryCode:        user.CountryCode,
		Mobile:             string(user.Mobile),
		Username:           user.Username,
		Email:              string(user.Email),
		Nickname:           user.Nickname,
		Avatar:             Url(user.Avatar),
		Bio:                user.Bio,
		CurrencyId:         user.CurrencyId,
		Signature:          UserSignature(user.ID),
		FollowingCount:     user.FollowingCount,
		HasSetSecondaryPwd: len(user.SecondaryPassword) > 0,
		Brand:              user.BrandId,
		Agent:              user.AgentId,
	}
	if user.Birthday.Valid {
		u.Birthday = user.Birthday.Time.Format(time.DateOnly)
	}
	if user.Username == "" {
		u.SetupRequired = true
	}
	if user.UserSum != nil {
		t := BuildUserSum(*user.UserSum)
		u.UserSum = &t
	}
	if user.Kyc != nil {
		t := BuildKyc(c, *user.Kyc, []model.KycDocument{})
		u.Kyc = &t
	} else {
		if user.KycCheckRequired && !user.SetupCompletedAt.IsZero() && user.SetupCompletedAt.Before(time.Now().Add(-7*24*time.Hour)) {
			u.KycRequired = true
		}
	}

	hasCompletedAchievement := map[int64]bool{}
	for _, a := range user.Achievements {
		hasCompletedAchievement[a.AchievementId] = true
	}
	u.HasCompletedFirstAppLoginTutorial = hasCompletedAchievement[model.UserAchievementIdFirstAppLoginTutorial]
	u.HasClaimedFirstAppLoginReward = hasCompletedAchievement[model.UserAchievementIdFirstAppLoginReward]

	return u
}

type SimpleUser struct {
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Bio      string `json:"bio,omitempty"`
}

func BuildSimpleUser(c *gin.Context, user model.User) SimpleUser {
	u := SimpleUser{
		Nickname: user.Nickname,
		Avatar:   Url(user.Avatar),
		Bio:      user.Bio,
	}
	return u
}

type UserAvatar struct {
	ID     int64  `json:"id"`
	Avatar string `json:"avatar"`
}

func BuildUserAvatar(c *gin.Context, user ploutos.User) UserAvatar {
	return UserAvatar{
		ID:     user.ID,
		Avatar: Url(user.Avatar),
	}
}
