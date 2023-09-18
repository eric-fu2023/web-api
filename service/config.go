package service

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
	"web-api/model"
	"web-api/serializer"
	"web-api/util/i18n"
)

type AppConfigService struct {
	Platform
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
	Platform
}

func (service *AnnouncementsService) Get(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	var announcements []ploutos.Announcement
	brand := c.MustGet(`_brand`).(int)
	agent := c.MustGet(`_agent`).(int)
	err = model.DB.Scopes(model.ByBrandAgentAndPlatform(int64(brand), int64(agent), service.Platform.Platform), model.Sort).Find(&announcements).Error
	if err != nil {
		r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}
	var texts []string
	for _, a := range announcements {
		texts = append(texts, a.Text)
	}
	r = serializer.Response{
		Data: texts,
	}
	return
}
