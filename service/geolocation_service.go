package service

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
	"web-api/cache"
	"web-api/model"
	"web-api/serializer"
	"web-api/util"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
)

const (
	RedisKeyGeolocation      = "geolocation:"
	RedisDurationGeolocation = 7 * 24 * 60 * 60 * time.Second
)

const (
	VENDOR_IP_API = 1
)

type GetGeolocationService struct{}

type GeolocationVendorResponse struct {
	As          string  `json:"as"`
	City        string  `json:"city"`
	Country     string  `json:"country"`
	CountryCode string  `json:"countryCode"`
	Isp         string  `json:"isp"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	Org         string  `json:"org"`
	Query       string  `json:"query"`
	Region      string  `json:"region"`
	RegionName  string  `json:"regionName"`
	Status      string  `json:"status"`
	Timezone    string  `json:"timezone"`
	Zip         string  `json:"zip"`
}

type GetGeolocationResponse struct {
	IpAddress string `json:"ip_address"`
	// ClientIpAddress string `json:"client_ip_address"`
	CountryCode string `json:"country_code"`
	// CountryName     string `json:"country_name,omitempty"`
	Region string `json:"region"`
	City   string `json:"city"`
	// Zipcode         string `json:"zipcode,omitempty"`
}

func (service *GetGeolocationService) Get(c *gin.Context) serializer.Response {
	// ip := retrieveClientIp(c)
	ip := c.ClientIP()

	// retrieve from Redis
	geolocation := retrieveGeolocationFromRedis(ip, c)
	// retrive from DB
	if len(geolocation.IpAddress) < 1 {
		geolocation = retrieveGeolocationFromDB(ip, c)
	}
	// retrive from vendor
	if len(geolocation.IpAddress) < 1 {
		geolocation = retrieveGeolocationFromVendor(ip, c)
	}

	return serializer.Response{
		Msg:  "success",
		Data: buildGeolocationResponse(ip, geolocation),
	}
}

func retrieveGeolocationFromRedis(ip string, c *gin.Context) ploutos.Geolocation {
	cacheData := cache.RedisGeolocationClient.Get(context.TODO(), RedisKeyGeolocation+ip)
	if cacheData.Err() != nil {
		return ploutos.Geolocation{}
	}

	geolocation := ploutos.Geolocation{}
	err := json.Unmarshal([]byte(cacheData.Val()), &geolocation)
	if err != nil {
		util.GetLoggerEntry(c).Warn("retrieveGeolocationFromRedis deserializing json failed: ", err.Error())
	}

	return geolocation
}

func retrieveGeolocationFromDB(ip string, c *gin.Context) ploutos.Geolocation {
	geolocation := model.FindGeolocation(ip, c)
	if geolocation.ID < 1 {
		return ploutos.Geolocation{}
	}

	// cache in Redis
	if jsonStr, err := json.Marshal(geolocation); err == nil {
		cache.RedisGeolocationClient.Set(context.TODO(), RedisKeyGeolocation+ip, jsonStr, RedisDurationGeolocation)
	} else {
		util.GetLoggerEntry(c).Warn("retrieveGeolocationFromDB serializing json failed: ", err.Error())
	}

	return geolocation
}

func retrieveGeolocationFromVendor(ip string, c *gin.Context) ploutos.Geolocation {
	// call vendor api
	client := http.Client{Timeout: 5 * time.Second}
	res, err := client.Get(strings.ReplaceAll(os.Getenv("GEOLOCATION_VENDOR_IPAPI_URL"), "{ip}", ip))
	if err != nil {
		util.GetLoggerEntry(c).Warn("queryGeolocationVendor query vendor api failed: ", err.Error())
		return ploutos.Geolocation{}
	}
	defer res.Body.Close()

	geolocation := buildGeolocation(ip, res, c)
	if len(geolocation.IpAddress) < 1 {
		return ploutos.Geolocation{}
	}

	err = model.DB.Save(&geolocation).Error
	if err != nil {
		util.GetLoggerEntry(c).Warn("retrieveGeolocationFromVendor save geolocation failed: ", err.Error())
	}
	if geolocation.ID > 0 {
		// cache in Redis
		if jsonStr, err := json.Marshal(geolocation); err == nil {
			cache.RedisGeolocationClient.Set(context.TODO(), RedisKeyGeolocation+ip, jsonStr, RedisDurationGeolocation)
		} else {
			util.GetLoggerEntry(c).Warn("retrieveGeolocationFromVendor serializing json failed: ", err.Error())
		}
	}
	return geolocation
}

func buildGeolocationResponse(ip string, geolocation ploutos.Geolocation) GetGeolocationResponse {
	res := GetGeolocationResponse{
		IpAddress: ip,
		// ClientIpAddress: clientIp,
		CountryCode: geolocation.CountryCode,
		Region:      geolocation.Region,
		City:        geolocation.City,
		// Zipcode:         geolocation.Zipcode,
	}
	// if countryCode := geolocation.CountryCode; len(countryCode) > 0 {
	// 	res.CountryName = consts.CountryMap[countryCode]
	// }
	return res
}

func buildGeolocation(ip string, res *http.Response, c *gin.Context) ploutos.Geolocation {
	body, err := io.ReadAll(res.Body)
	if err != nil {
		util.GetLoggerEntry(c).Warn("queryGeolocationVendor read body failed: ", err.Error())
		return ploutos.Geolocation{}
	}

	geolocationVendorResponse := GeolocationVendorResponse{}
	err = json.Unmarshal(body, &geolocationVendorResponse)
	if err != nil {
		util.GetLoggerEntry(c).Warn("queryGeolocationVendor deserialize json failed: ", err.Error())
		return ploutos.Geolocation{}
	}

	geolocation := ploutos.Geolocation{
		IpAddress:   ip,
		CountryCode: geolocationVendorResponse.CountryCode,
		Vendor:      VENDOR_IP_API,
		VendorData:  body,
	}
	if region := geolocationVendorResponse.RegionName; len(region) > 0 {
		geolocation.Region = region
	}
	if city := geolocationVendorResponse.City; len(city) > 0 {
		geolocation.City = city
	}
	if zipcode := geolocationVendorResponse.Zip; len(zipcode) > 0 {
		geolocation.Zipcode = zipcode
	}

	return geolocation
}
