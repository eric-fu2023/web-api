package model

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"time"
	"web-api/util"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrCannotFindUser = errors.New("cannot find user")
)

type User struct {
	ploutos.User
	Mobile           ploutos.EncryptedStr `json:"mobile" form:"mobile" gorm:"column:mobile"`
	Email            ploutos.EncryptedStr `json:"email" form:"email" gorm:"column:email"`
	KycCheckRequired bool                 `gorm:"-"`
	UserSum          *ploutos.UserSum     `gorm:"foreignKey:UserId;references:ID"`
	Kyc              *Kyc                 `gorm:"foreignKey:UserId;references:ID"`
	Achievements     []UserAchievement    `gorm:"foreignKey:UserId;references:ID"`
}

const (
	PassWordCost        = 4
	Active       string = "active"
	Inactive     string = "inactive"
	Suspend      string = "suspend"
)

func (user *User) IdAsString() string {
	return strconv.FormatInt(user.ID, 10)
}

func (user *User) GenToken() (tokenString string, err error) {
	days, _ := strconv.Atoi(os.Getenv("TOKEN_EXPIRED_DAYS"))
	now := time.Now()
	exp := now.AddDate(0, 0, days)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":      user.ID,
		"country_code": user.CountryCode,
		"mobile":       string(user.Mobile),
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

func (user *User) UpdateLoginInfo(c *gin.Context, loginTime time.Time) error {
	deviceInfo, err := util.GetDeviceInfo(c)
	if err != nil {
		util.GetLoggerEntry(c).Errorf("GetDeviceInfo error: %s", err.Error())
		return util.ErrInvalidDeviceInfo
	}

	update := User{
		User: ploutos.User{
			BASE:                ploutos.BASE{ID: user.ID},
			LastLoginIp:         c.ClientIP(),
			LastLoginTime:       loginTime,
			LastLoginDeviceUuid: deviceInfo.Uuid, // Not updated if empty
		},
	}

	if err = DB.Updates(update).Error; err != nil {
		util.GetLoggerEntry(c).Errorf("Update last login info error: %s", err.Error())
		return err
	}

	return nil
}

func (user *User) CreateWithDB(tx *gorm.DB) error {
	if err := tx.Create(&user).Error; err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	user.ReferralCode = generateReferralCode(user.ID)
	if err := tx.Updates(&user).Error; err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	return nil
}

const (
	// Note: Do not modify
	alp1 = "GLKZT4B98Y7P6ENXWU32M5VCQSFDHRJA"
	alp2 = "JC5MHVLGRFP4UBWNTZ6D9X8Y3QKAS7E2"
	alp3 = "DMG7592UV4XL86CPSEWTK3NBRZHQFYJA"
	alp4 = "YV8DFX4PERBAZ629T5HCK3JMUQWN7LGS"
	alp5 = "KBQAEFJ6X4M5VDC2NG7LU3ZPRH98TYWS"
	alp6 = "ZANWSM9JV6HE4LY5DUC87PGT32XFBRQK"
)

func generateReferralCode(userId int64) string {
	referralCode := ""

	referralCode += string(alp1[(userId/int64(math.Pow(float64(32), float64(5))))%32])
	referralCode += string(alp2[(userId/int64(math.Pow(float64(32), float64(4))))%32])
	referralCode += string(alp3[(userId/int64(math.Pow(float64(32), float64(3))))%32])
	referralCode += string(alp4[(userId/int64(math.Pow(float64(32), float64(2))))%32])
	referralCode += string(alp5[(userId/int64(math.Pow(float64(32), float64(1))))%32])
	referralCode += string(alp6[(userId/int64(math.Pow(float64(32), float64(0))))%32])

	return referralCode
}

// GetUserLang defaults to env config. see commented out for prev impl.
func GetUserLang(_ int64) (string, error) {
	v := os.Getenv("PLATFORM_LANGUAGE")
	if v == "" {
		return "", errors.New("platform.language default not set")
	}
	return v, nil
	//var user User
	//err := DB.First(&user).Error
	//if err != nil {
	//	return "", err
	//}
	//
	//if len(user.Locale) < 2 {
	//	return "", errors.New("invalid locale of minimum length 2")
	//}
	//return user.Locale[:2], nil
}

func GetUserByMobileOrEmailOld(countryCode, mobile, email string) (User, error) {
	var user User

	if mobile != "" && countryCode != "" {
		mobileHash := util.MobileEmailHash(mobile)
		if err := DB.Where(`country_code`, countryCode).Where(`mobile_hash`, mobileHash).First(&user).Error; err != nil {
			return User{}, err
		}
	} else if email != "" {
		emailHash := util.MobileEmailHash(email)
		if err := DB.Where(`email_hash`, emailHash).First(&user).Error; err != nil {
			return User{}, err
		}
	} else {
		return User{}, ErrCannotFindUser
	}

	return user, nil
}

func GetUserByMobileOrEmail(countryCode, mobile, email string) (User, error) {
	var users []User
	if mobile != "" && countryCode != "" {
		mobileHash := util.MobileEmailHash(mobile)
		if err := DB.Where(`country_code`, countryCode).Where(`mobile_hash`, mobileHash).Find(&users).Error; err != nil {
			return User{}, err
		}
	}
	if len(users) > 0 {
		return users[0], nil
	}

	if email != "" {
		emailHash := util.MobileEmailHash(email)
		if err := DB.Where(`email_hash`, emailHash).First(&users).Error; err != nil {
			return User{}, err
		}
	}
	if len(users) > 0 {
		return users[0], nil
	}

	return User{}, ErrCannotFindUser
}

// IP防薅
func IPExisted(ip string) (isExisted bool) {
	var count int64
	err := DB.Table("users").
		Where("registration_ip = ?", ip).
		Count(&count).Error

	if err != nil || count > 1 {
		isExisted = true
	}

	return
}
