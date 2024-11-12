package service

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"web-api/cache"
	"web-api/model"
	"web-api/serializer"
	"web-api/service/common"
	"web-api/util"
	"web-api/util/i18n"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
)

type AppConfigService struct {
	common.Platform
	Key string `form:"key" json:"key"`
}

func (service *AppConfigService) Get(c *gin.Context) (r serializer.Response, err error) {

	// log install event when api calls for config
	agent := c.GetHeader("channel")
	if agent == "pixel_app_001"{
		log.Printf("should log pixel event view content for channel pixel_app_001")
		PixelInstallEvent(c.ClientIP())
	}

	// retrieve basic AppConfigs
	cf, err := service.getAppConfigs(c)
	if err != nil {
		util.GetLoggerEntry(c).Errorf("getAppConfigs err: %s", err.Error())
		r = serializer.GeneralErr(c, err)
		return
	}

	// Get AB toggle configs
	isA, err := service.isA(c)
	if err != nil {
		util.GetLoggerEntry(c).Errorf("isA err: %s", err.Error())
		r = serializer.GeneralErr(c, err)
		return
	}

	// If is A, replace chat socket url with hardcoded value
	socketChatUrl := os.Getenv("SOCKET_CHAT_URL_A")
	if isA && socketChatUrl != "" {
		cf["socket"]["chat_url"] = socketChatUrl
	}
	cf["ab"] = map[string]string{
		"is_a": strconv.FormatBool(isA),
	}

	// retrieve channelCode based on origin domain
	var domainWebConfigService DomainWebConfigService
	cf["channel"] = map[string]string{
		"code": domainWebConfigService.RetrieveChannel(c),
	}

	r = serializer.Response{
		Data: cf,
	}

	return
}

// getAppConfigs retrieves the app config from cache or db
func (service *AppConfigService) getAppConfigs(c *gin.Context) (cf map[string]map[string]string, err error) {
	brand := c.MustGet(`_brand`).(int)
	cf, err = cache.GetAppConfig(brand, service.Platform.Platform, service.Key)
	if err != nil && !errors.Is(err, cache.ErrCacheMiss) {
		// Don't block request if cache fails, just log the error and query db
		util.GetLoggerEntry(c).Errorf("GetAppConfig: %s", err.Error())
	}
	if err == nil {
		return cf, nil
	}

	var configs []ploutos.AppConfig
	err = model.DB.Scopes(model.ByBrandPlatformAndKey(int64(brand), service.Platform.Platform, service.Key)).Find(&configs).Error
	if err != nil {
		return nil, fmt.Errorf("find app configs from db: %w", err)
	}

	cf = make(map[string]map[string]string)
	for _, b := range configs {
		_, exists := cf[b.Name]
		if !exists {
			cf[b.Name] = make(map[string]string)
		}
		cf[b.Name][b.Key] = b.Value
	}

	err = cache.SetAppConfig(brand, service.Platform.Platform, service.Key, cf)
	if err != nil {
		// Don't block request if cache fails, just log the error and return
		util.GetLoggerEntry(c).Errorf("SetAppConfig: %s", err.Error())
	}

	return cf, nil
}

func (service *AppConfigService) isA(c *gin.Context) (isA bool, err error) {
	deviceInfo, err := util.GetDeviceInfo(c)
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
	if service.Version == "1.0.11" {
		return
	}
	i18n := c.MustGet("i18n").(i18n.I18n)
	brandId := c.MustGet("_brand").(int)
	channel := c.MustGet("_channel").(string)
	var app model.AppUpdate
	err = app.Get(int64(brandId), service.Platform, service.Version, channel)
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
