package service

import (
	"context"
	"encoding/json"
	"time"
	"web-api/cache"
	"web-api/model"
	"web-api/serializer"
	"web-api/util"

	"github.com/gin-gonic/gin"
)

const (
	RedisKeyDomainApiConfigs = "domain_api_configs:"
)

type DomainConfigService struct{}

type DomainConfigRes struct {
	Api    string `json:"a"`
	Record string `json:"r"`
	Nami   string `json:"n"`
}

func (service *DomainConfigService) InitApp(c *gin.Context) (code int, res serializer.Response, err error) {
	// TODO: check if the request comes from countries that should be blocked
	shouldBlock := false
	if shouldBlock {
		code = serializer.CodeNoRightErr
		res = serializer.Response{
			Msg: "blocked",
		}
		return
	}

	// retrieve all active app domains, shuffle and return
	code = 200
	res = serializer.Response{
		Msg: "success",
		Data: DomainConfigRes{
			Api:    FindRandomDomain(model.SupportTypeApp, model.DomainTypeApi, c),
			Record: FindRandomDomain(model.SupportTypeApp, model.DomainTypeRecord, c),
			Nami:   FindRandomDomain(model.SupportTypeApp, model.DomainTypeNami, c),
		},
	}

	return
}

func FindRandomDomain(supportType string, domainType string, c *gin.Context) string {
	// retrieve from Redis
	if domain := retrieveFromRedis(supportType, domainType, c); len(domain) > 0 {
		return domain
	}
	// retrieve from DB & cache into Redis
	if domain := retrieveFromDB(supportType, domainType, c); len(domain) > 0 {
		return domain
	}
	return ""
}

func retrieveFromRedis(supportType string, domainType string, c *gin.Context) string {
	cacheData := cache.RedisDomainConfigClient.Get(context.TODO(), RedisKeyDomainApiConfigs+supportType)
	if cacheData.Err() != nil {
		return ""
	}

	// deserialize into map
	var domainsByType = make(map[string][]model.DomainApiConfig)
	err := json.Unmarshal([]byte(cacheData.Val()), &domainsByType)
	if err != nil {
		util.GetLoggerEntry(c).Warn("FindRandomDomain deserializing json failed: ", err.Error())
		return ""
	}

	// pick pseudo-randomly
	domains := domainsByType[domainType]
	size := len(domains)
	if size > 0 {
		return domains[time.Now().UTC().UnixMicro()%int64(size)].DomainUrl
	}
	return ""
}

func retrieveFromDB(supportType string, domainType string, c *gin.Context) string {
	domains := model.DomainApiConfig{}.FindDomainConfigs(supportType, c)
	if len(domains) < 1 {
		return ""
	}

	// convert to map with key = DomainType
	domainsByType := make(map[string][]model.DomainApiConfig)
	for _, domain := range domains {
		domainsByType[domain.DomainType] = append(domainsByType[domain.DomainType], domain)
	}

	// cache in Redis
	if jsonStr, err := json.Marshal(domainsByType); err == nil {
		cache.RedisDomainConfigClient.Set(context.TODO(), RedisKeyDomainApiConfigs+supportType, jsonStr, 3*60*time.Second)
	} else {
		util.GetLoggerEntry(c).Warn("retrieveFromDB serializing json failed: ", err.Error())
	}

	// pick pseudo-randomly
	domains = domainsByType[domainType]
	size := len(domains)
	if size > 0 {
		return domains[time.Now().UTC().UnixMicro()%int64(size)].DomainUrl
	}
	return ""
}
