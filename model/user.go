package model

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
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

func GetUserLang(userId int64) string {
	var user User
	DB.First(&user)
	return user.Locale[:2]
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

var avatars = []string{
	"https://cdn.tayalive.com/aha-img/user/default_user_image/100.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/101.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/102.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/103.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/104.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/105.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/107.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/108.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/109.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/110.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/111.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/112.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/113.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/114.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/115.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/116.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/117.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/118.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/119.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/120.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/121.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/122.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/123.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/124.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/125.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/126.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/127.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/128.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/129.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/130.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/131.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/132.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/133.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/134.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/135.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/136.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/137.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/138.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/139.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/140.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/141.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/142.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/143.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/144.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/145.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/146.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/147.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/148.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/149.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/150.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/151.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/152.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/153.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/154.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/155.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/156.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/157.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/21.jpeg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/22.jpeg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/23.jpeg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/24.jpeg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/24.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/25.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/26.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/27.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/28.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/29.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/30.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/31.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/32.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/33.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/34.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/35.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/36.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/37.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/38.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/39.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/40.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/41.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/42.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/43.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/44.jpeg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/45.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/46.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/47.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/48.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/49.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/50.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/51.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/52.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/53.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/54.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/55.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/56.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/57.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/58.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/59.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/60.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/61.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/62.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/63.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/64.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/65.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/66.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/67.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/68.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/69.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/70.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/71.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/72.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/73.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/74.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/75.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/76.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/77.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/78.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/79.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/8.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/80.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/81.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/82.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/83.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/84.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/85.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/86.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/87.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/88.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/89.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/90.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/91.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/92.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/93.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/94.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/95.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/96.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/97.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/98.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/99.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/p1.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/p10.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/p11.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/p12.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/p13.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/p14.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/p15.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/p16.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/p17.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/p2.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/p3.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/p4.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/p5.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/p6.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/p7.jpg",
	"https://cdn.tayalive.com/aha-img/user/default_user_image/p9.jpg",
}

func SetRandomAvatar(user *User) {
	rand := rand.Intn(len(avatars))
	user.Avatar = avatars[rand]
}
