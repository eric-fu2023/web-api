package service

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"mime/multipart"
	"time"
	"web-api/model"
	"web-api/serializer"
	"web-api/util"
	"web-api/util/i18n"
)

type ProfileUpdateService struct {
	Nickname       string                `form:"nickname" json:"nickname"`
	ProfilePicture *multipart.FileHeader `form:"profile_picture" json:"profile_picture"`
	Birthday       string                `form:"birthday" json:"birthday"` // YYYY-MM-DD
}

func (service *ProfileUpdateService) Update(c *gin.Context) serializer.Response {
	u, _ := c.Get("user")
	user := u.(model.User)

	validateResp := service.validate(c, user)
	if validateResp.Code != 0 {
		return validateResp
	}

	err := service.updateUser(&user)
	if err != nil {
		util.GetLoggerEntry(c).Errorf("update user error: %s", err.Error())
		return serializer.GeneralErr(c, err)
	}

	userAchievements, err := model.GetUserAchievementsForMe(user.ID)
	if err != nil {
		util.GetLoggerEntry(c).Errorf("get user achievements error: %s", err.Error())
		return serializer.GeneralErr(c, err)
	}
	user.Achievements = userAchievements

	return serializer.Response{
		Data: serializer.BuildUserInfo(c, user),
	}
}

func (service *ProfileUpdateService) validate(c *gin.Context, user model.User) serializer.Response {
	i18n := c.MustGet("i18n").(i18n.I18n)

	if service.Birthday != "" {
		_, err := time.Parse(time.DateOnly, service.Birthday)
		if err != nil {
			return serializer.ParamErr(c, service, fmt.Sprintf(i18n.T("kyc_invalid_birthday"), service.Birthday), err)
		}

		// Check if user has already updated birthday before
		uas, err := model.GetUserAchievements(user.ID, model.GetUserAchievementCond{AchievementIds: []int64{model.UserAchievementIdUpdateBirthday}})
		if err != nil {
			util.GetLoggerEntry(c).Errorf("get user achievement error: %s", err.Error())
			return serializer.DBErr(c, service, i18n.T("general_error"), err)
		}
		if len(uas) > 0 {
			return serializer.ParamErr(c, service, i18n.T("user_update_birthday_once"), err)
		}
	}

	if service.ProfilePicture != nil {
		mt := service.ProfilePicture.Header.Get("Content-Type")
		if _, exists := profilePicContentTypeToExt[mt]; !exists {
			return serializer.ParamErr(c, service, fmt.Sprintf(i18n.T("file_type_not_allowed"), mt), nil)
		}
		if service.ProfilePicture.Size > fileSizeLimit*1024*1024 {
			return serializer.ParamErr(c, service, fmt.Sprintf(i18n.T("file_size_exceeded"), fmt.Sprintf(`%d MB`, fileSizeLimit)), nil)
		}
	}

	return serializer.Response{}
}

func (service *ProfileUpdateService) updateUser(user *model.User) (err error) {
	err = model.DB.Transaction(func(tx *gorm.DB) error {
		if service.Nickname != "" {
			user.Nickname = service.Nickname
		}

		if service.Birthday != "" {
			birthday, _ := time.Parse(time.DateOnly, service.Birthday)
			if user.Birthday.Valid { // if user has updated birthday before
				err := model.CreateUserAchievementWithDB(tx, user.ID, model.UserAchievementIdUpdateBirthday)
				if err != nil {
					return fmt.Errorf("create birthday achievement: %w", err)
				}
			}
			user.Birthday = sql.NullTime{Time: birthday, Valid: true}
		}

		if service.ProfilePicture != nil {
			path, err := service.uploadProfilePic(*user)
			if err != nil {
				return fmt.Errorf("upload profile pic: %w", err)
			}
			user.Avatar = path
		}

		err = tx.Updates(user).Error
		if err != nil {
			return fmt.Errorf("update user: %w", err)
		}

		return nil
	})

	return
}

func (service *ProfileUpdateService) uploadProfilePic(user model.User) (path string, err error) {
	oss, err := util.InitAliyunOSS()
	if err != nil {
		return "", fmt.Errorf("initAliyunOSS: %w", err)
	}

	mt := service.ProfilePicture.Header.Get("Content-Type")
	path, err = oss.UploadFile(
		util.AliyunOssAvatar,
		user.ID,
		service.ProfilePicture,
		profilePicContentTypeToExt[mt])
	if err != nil {
		return "", fmt.Errorf("uploadFile: %w", err)
	}

	return path, err
}

// Deprecated: Use ProfileUpdateService instead
type NicknameUpdateService struct {
	Nickname *string `form:"nickname" json:"nickname"`
}

func (service *NicknameUpdateService) Update(c *gin.Context) serializer.Response {
	i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")
	user := u.(model.User)

	user.Nickname = *service.Nickname
	err := model.DB.Model(model.User{}).Where(`id`, user.ID).Update(`nickname`, user.Nickname).Error
	if err != nil {
		return serializer.DBErr(c, service, i18n.T("general_error"), err)
	}

	return serializer.Response{
		Data: serializer.BuildUserInfo(c, user),
	}
}

// Deprecated: Use ProfileUpdateService instead
type ProfilePicService struct {
	File multipart.FileHeader `form:"file" json:"file" binding:"gt=0"`
}

var profilePicContentTypeToExt = map[string]string{
	"image/jpeg": ".jpg",
	"image/jpg":  ".jpg",
	"image/png":  ".png",
}
var fileSizeLimit = int64(5) // MB

func (service *ProfilePicService) Upload(c *gin.Context) serializer.Response {
	i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")
	user := u.(model.User)

	mt := service.File.Header.Get("Content-Type")
	if _, exists := profilePicContentTypeToExt[mt]; !exists {
		return serializer.ParamErr(c, service, fmt.Sprintf(i18n.T("file_type_not_allowed"), mt), nil)
	}
	if service.File.Size > fileSizeLimit*1024*1024 {
		return serializer.ParamErr(c, service, fmt.Sprintf(i18n.T("file_size_exceeded"), fmt.Sprintf(`%d MB`, fileSizeLimit)), nil)
	}

	oss, err := util.InitAliyunOSS()
	if err != nil {
		return serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
	}
	path, err := oss.UploadFile(
		util.AliyunOssAvatar,
		user.ID,
		&service.File,
		profilePicContentTypeToExt[mt])
	if err != nil {
		return serializer.DBErr(c, service, i18n.T("general_error"), err)
	}

	user.Avatar = path
	err = model.DB.Model(model.User{}).Where(`id`, user.ID).Update(`avatar`, path).Error
	if err != nil {
		return serializer.DBErr(c, service, i18n.T("general_error"), err)
	}

	return serializer.Response{
		Data: serializer.BuildUserInfo(c, user),
	}
}
