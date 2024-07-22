package service

import (
	"fmt"
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
		Vendor:      ploutos.IP_API,
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

type GetGeolocationService struct {
	Key string `form:"key" json:"key"`
}

func (service *GetGeolocationService) Get(c *gin.Context) serializer.Response {
	for k, v := range c.Request.Header {
		fmt.Printf("Header field %q, Value %q\n", k, v)
	}
	fmt.Printf("Header field %q, Value %q\n", "remoteAddress", c.Request.RemoteAddr)
	return serializer.Response{
		Msg: "success",
	}
}
