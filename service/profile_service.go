package service

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"mime/multipart"
	"time"
	"web-api/model"
	"web-api/serializer"
	"web-api/util"
	"web-api/util/i18n"
)

type ProfileGetService struct {
}

func (service *ProfileGetService) Get(c *gin.Context) serializer.Response {
	i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")
	user := u.(model.User)

	userProfile, err := getUserProfile(user)
	if err != nil {
		return serializer.DBErr(c, service, i18n.T("general_error"), err)
	}

	return serializer.Response{
		Data: serializer.BuildProfile(c, userProfile),
	}
}

type ProfileUpdateService struct {
	Nickname   *string `form:"nickname" json:"nickname"`
	FirstName  *string `form:"first_name" json:"first_name"`
	MiddleName *string `form:"middle_name" json:"middle_name"`
	LastName   *string `form:"last_name" json:"last_name"`
	Street     *string `form:"street" json:"street"`
	Province   *string `form:"province" json:"province"`
	City       *string `form:"city" json:"city"`
	Postcode   *string `form:"postcode" json:"postcode"`
	Birthday   *string `form:"birthday" json:"birthday"`
}

func (service *ProfileUpdateService) Update(c *gin.Context) serializer.Response {
	i18n := c.MustGet("i18n").(i18n.I18n)
	var err error
	var birthday time.Time
	if service.Birthday != nil {
		birthday, err = time.Parse(time.DateOnly, *service.Birthday)
		if err != nil {
			return serializer.ParamErr(c, service, fmt.Sprintf(i18n.T("kyc_invalid_birthday"), service.Birthday), err)
		}
	}

	u, _ := c.Get("user")
	user := u.(model.User)

	userProfile, err := getUserProfile(user)
	if err != nil {
		return serializer.DBErr(c, service, i18n.T("general_error"), err)
	}

	copier.Copy(&userProfile, &service)
	userProfile.Birthday = birthday
	err = model.DB.Save(&userProfile).Error
	if err != nil {
		return serializer.DBErr(c, service, i18n.T("general_error"), err)
	}

	return serializer.Response{
		Data: serializer.BuildProfile(c, userProfile),
	}
}

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

	userProfile, err := getUserProfile(user)
	if err != nil {
		return serializer.DBErr(c, service, i18n.T("general_error"), err)
	}

	return serializer.Response{
		Data: serializer.BuildProfile(c, userProfile),
	}
}

func getUserProfile(user model.User) (userProfile model.UserProfile, err error) {
	userProfile.User = &user
	userProfile.UserId = user.ID
	err = model.DB.Scopes(model.ByUserId(user.ID)).Find(&userProfile).Error
	return
}
