package service

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"mime/multipart"
	"web-api/model"
	"web-api/serializer"
	"web-api/util"
	"web-api/util/i18n"
)

var (
	feedbackFileContentTypeToExtension = map[string]string{
		"image/jpeg": ".jpg",
		"image/jpg":  ".jpg",
		"image/png":  ".png",
	}
	feedbackFileSizeLimit int64 = 5 * 1024 * 1024 // 5MB
)

type FeedbackAddService struct {
	Content string                 `form:"content" json:"content" binding:"required"`
	Images  []multipart.FileHeader `form:"images" json:"images"`
}

func (service *FeedbackAddService) Add(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")
	user := u.(model.User)

	oss, err := util.InitAliyunOSS()
	if err != nil {
		util.GetLoggerEntry(c).Errorf("Init oss error: %s", err.Error())
		return
	}

	var paths []string
	for _, image := range service.Images {
		mt := image.Header.Get("Content-Type")
		if _, exists := feedbackFileContentTypeToExtension[mt]; !exists {
			r = serializer.ParamErr(c, service, fmt.Sprintf(i18n.T("file_type_not_allowed"), mt), nil)
			return
		}
		if image.Size > feedbackFileSizeLimit {
			serializer.ParamErr(c, service, fmt.Sprintf(i18n.T("file_size_exceeded"), "2MB"), nil)
			return
		}
		path, e := oss.UploadFile(util.AliyunOssFeedback, user.ID, &image, feedbackFileContentTypeToExtension[image.Header.Get("Content-Type")])
		if e != nil {
			r = serializer.ParamErr(c, service, i18n.T("file_upload_error"), e)
			go deleteOSSFiles(paths)
			return
		}
		paths = append(paths, path)
	}

	feedback := ploutos.Feedback{
		UserId:  user.ID,
		Content: service.Content,
	}
	tx := model.DB.Begin()
	err = tx.Save(&feedback).Error
	if err != nil {
		r = serializer.Err(c, "", serializer.CodeGeneralError, i18n.T("general_error"), err)
		tx.Rollback()
		return
	}
	for _, path := range paths {
		image := ploutos.FeedbackImage{
			FeedbackId: feedback.ID,
			Url:        path,
		}
		err = tx.Save(&image).Error
		if err != nil {
			r = serializer.Err(c, "", serializer.CodeGeneralError, i18n.T("general_error"), err)
			tx.Rollback()
			return
		}
	}
	tx.Commit()

	r = serializer.Response{
		Msg: i18n.T("success"),
	}
	return
}

func deleteOSSFiles(paths []string) {
	oss, err := util.InitAliyunOSS()
	if err != nil {
		util.GetLoggerEntry(context.TODO()).Errorf("OSS delete file error: %s", err.Error())
	}
	for _, path := range paths {
		e := oss.DeleteFile(path)
		if e != nil {
			util.GetLoggerEntry(context.TODO()).Errorf("OSS delete file error: %s", e.Error())
		}
	}
}
