package service

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"strconv"
	"web-api/model"
	"web-api/serializer"
	"web-api/service/common"
	"web-api/util"
	"web-api/util/i18n"
)

type AppConfigService struct {
	common.Platform
	Key string `form:"key" json:"key"`
}

func (service *AppConfigService) Get(c *gin.Context) (r serializer.Response, err error) {
	var configs []ploutos.AppConfig
	brand := c.MustGet(`_brand`).(int)
	//agent := c.MustGet(`_agent`).(int)

	err = model.DB.Scopes(model.ByBrandPlatformAndKey(int64(brand), service.Platform.Platform, service.Key)).Find(&configs).Error
	if err != nil {
		r = serializer.GeneralErr(c, err)
	}

	cf := make(map[string]map[string]string)
	for _, b := range configs {
		_, exists := cf[b.Name]
		if !exists {
			cf[b.Name] = make(map[string]string)
		}
		cf[b.Name][b.Key] = b.Value
	}

	// Get AB toggle configs
	isA, err := service.isA(c)
	if err != nil {
		util.GetLoggerEntry(c).Errorf("isA err: %s", err.Error())
		r = serializer.GeneralErr(c, err)
	}

	cf["ab"] = map[string]string{
		"is_a": strconv.FormatBool(isA),
	}

	r = serializer.Response{
		Data: cf,
	}

	return
}

func (service *AppConfigService) isA(c *gin.Context) (isA bool, err error) {
	deviceInfo, err := util.GetDeviceInfo(c)
	err = errors.New("potato err")
	if err != nil {
		err = fmt.Errorf("getDeviceInfo: %w", err)
		return
	}

	sources := []string{c.ClientIP()}
	if deviceInfo.Uuid != "" {
		sources = append(sources, deviceInfo.Uuid)
	}
	if deviceInfo.Version != "" {
		sources = append(sources, deviceInfo.Version)
	}

	var toggles []ploutos.AbToggle
	err = model.DB.Where("is_a", true).Where("source IN ?", sources).Find(&toggles).Error
	if err != nil {
		err = fmt.Errorf("find ab toggles: %w", err)
		return
	}

	if len(toggles) > 0 {
		isA = true
	}
	return
}

type AnnouncementsService struct {
	common.Platform
}

func (service *AnnouncementsService) Get(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	var announcements []ploutos.Announcement
	_, isUser := c.Get("user")
	brand := c.MustGet(`_brand`).(int)
	//agent := c.MustGet(`_agent`).(int)
	loginStatus := 1
	if isUser {
		loginStatus = 2
	}
	err = model.DB.Scopes(model.ByBrandAndPlatform(int64(brand), service.Platform.Platform), model.ByTimeRange, model.ByStatus, model.SortAsc).
		Where(`login_status = 0 OR login_status = ?`, loginStatus).Find(&announcements).Error
	if err != nil {
		r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}
	r = serializer.Response{
		Data: serializer.BuildAnnouncements(announcements),
	}
	return
}

type AppUpdateService struct {
	Platform int64  `form:"platform" json:"platform" binding:"required"`
	Version  string `form:"version" json:"version" binding:"required"`
}

func (service *AppUpdateService) Get(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	brandId := c.MustGet("_brand").(int)
	var app model.AppUpdate
	err = app.Get(int64(brandId), service.Platform, service.Version)
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
