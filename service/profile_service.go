package service

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"mime/multipart"
	"web-api/model"
	"web-api/serializer"
	"web-api/util"
	"web-api/util/i18n"
)

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
