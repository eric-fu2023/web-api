package service

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/serializer"
	"web-api/service/common"
	"web-api/util/i18n"
)

type AppConfigService struct {
	common.Platform
	Key string `form:"key" json:"key"`
}

func (service *AppConfigService) Get(c *gin.Context) (r serializer.Response, err error) {
	var configs []ploutos.AppConfig
	brand := c.MustGet(`_brand`).(int)
	agent := c.MustGet(`_agent`).(int)
	if err = model.DB.Scopes(model.ByBrandAgentPlatformAndKey(int64(brand), int64(agent), service.Platform.Platform, service.Key)).Find(&configs).Error; err == nil {
		cf := make(map[string]map[string]string)
		for _, b := range configs {
			_, exists := cf[b.Name]
			if !exists {
				cf[b.Name] = make(map[string]string)
			}
			cf[b.Name][b.Key] = b.Value
		}
		r = serializer.Response{
			Data: cf,
		}
	}
	return
}

type AnnouncementsService struct {
	common.Platform
}

func (service *AnnouncementsService) Get(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	var announcements []ploutos.Announcement
	brand := c.MustGet(`_brand`).(int)
	agent := c.MustGet(`_agent`).(int)
	err = model.DB.Scopes(model.ByBrandAgentAndPlatform(int64(brand), int64(agent), service.Platform.Platform), model.ByMaintenance, model.ByStatus, model.Sort).Find(&announcements).Error
	if err != nil {
		r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}
	var texts, images []string
	for _, a := range announcements {
		if a.Type == consts.AnnouncementType["text"] {
			texts = append(texts, a.Text)
		} else if a.Type == consts.AnnouncementType["image"] {
			images = append(images, serializer.Url(a.Image))
		}
	}
	r = serializer.Response{
		Data: map[string][]string{
			"texts":  texts,
			"images": images,
		},
	}
	return
}

type AppUpdateService struct {
	Platform int64  `form:"platform" json:"platform" binding:"required"`
	Channel  int64  `form:"channel" json:"channel" binding:"required"`
	Version  string `form:"version" json:"version" binding:"required"`
}

func (service *AppUpdateService) Get(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	brandId := c.MustGet("_brand").(int)
	var app model.AppUpdate
	err = app.Get(int64(brandId), service.Platform, service.Channel, service.Version)
	if err != nil {
		r = serializer.Err(c, "", serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}
	var data *serializer.AppUpdate
	if app.ID != 0 {
		d := serializer.BuildAppUpdate(app)
		data = &d
	}
	r = serializer.Response{
		Data: data,
	}
	return
}
