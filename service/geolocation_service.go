package service

import (
	"log"
	"strings"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/serializer"
	"web-api/util"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm/clause"
)

const (
	RedisKeyGeolocation = "geolocation:"
)

type CreateGeolocationService struct {
	IpAddress   string `form:"ip_address" json:"ip_address" binding:"required"`
	CountryCode string `form:"country_code" json:"country_code" binding:"required"`
	Region      string `form:"region" json:"region"`
	City        string `form:"city" json:"city"`
	Zipcode     string `form:"zipcode" json:"zipcode"`
}

func (service *CreateGeolocationService) Create(c *gin.Context) serializer.Response {
	geolocation := ploutos.Geolocation{
		IpAddress:   service.IpAddress,
		CountryCode: service.CountryCode,
		Region:      service.Region,
		City:        service.City,
		Zipcode:     service.Zipcode,
		// Vendor:      ploutos.IP_API,
	}

	var err = model.DB.Clauses(clause.OnConflict{DoNothing: true}).Create(&geolocation).Error
	if err != nil {
		util.GetLoggerEntry(c).Errorf("Failed to create Geolocation: %s", err.Error())
		return serializer.GeneralErr(c, err)
	}

	return serializer.Response{
		Msg: "success",
	}
}

type GetGeolocationService struct{}

type Geolocation struct {
	IpAddress   string `json:"ip_address"`
	CountryCode string `json:"country_code"`
	CountryName string `json:"country_name"`
	Region      string `json:"region,omitempty"`
	City        string `json:"city,omitempty"`
	Zipcode     string `json:"zipcode,omitempty"`
}

func (service *GetGeolocationService) Get(c *gin.Context) serializer.Response {
	geolocation := Geolocation{
		IpAddress:   retrieveClientIp(c),
		CountryCode: "SG",
		CountryName: "Singapore",
		Region:      "Singapore",
		City:        "Singapore",
		Zipcode:     "999999",
	}

	return serializer.Response{
		Msg:  "success",
		Data: geolocation,
	}
}

func retrieveClientIp(c *gin.Context) string {
	clientIpStr := c.Request.Header.Get(consts.ClientIpHeader)
	if len(clientIpStr) > 0 {
		clientIps := strings.Split(clientIpStr, ",")
		return strings.TrimSpace(clientIps[0])
	}
	// fallback IP
	log.Printf("Cannot retrieve %s from header, using fallback IP instead", consts.ClientIpHeader)
	return "219.75.27.16"
}
