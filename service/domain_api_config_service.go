package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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
	// for both ba and aha, web-api
	Api string `json:"a,omitempty"`
	// for both ba and aha, record-center
	Record string `json:"r,omitempty"`
	// for aha only, hisport api to retrieve nami data
	Nami string `json:"n,omitempty"`
	// for ba only, batace api
	BataceApi string `json:"b,omitempty"`
	// for ba only, crickong api
	CrickongApi string `json:"c,omitempty"`
	// for ba only, A screen or B
	Mode bool `json:"m"`
	// mode from Taiwan Team
	TaiwanMode TaiwanModeResponse `json:"t_m"`
}

type CallbackSecretData struct {
	SecretData string `json:"secretData"`
}

type TaiwanModeResponse struct {
	IsValid   bool     `json:"isValid"`
	UpdateUrl string   `json:"updateUrl"`
	BaseUrls  []string `json:"BaseUrls"`
	WssUrl    string   `json:"WssUrl"`
	ImgUrl    string   `json:"ImgUrl"`
	Ip        string   `json:"Ip"`
	Location  string   `json:"Location"`
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

	queryParameters := generateParametersByHeaderIsAB(c)

	var twResp TaiwanModeResponse
	if len(queryParameters) != 0 {
		jsonData, _ := json.Marshal(queryParameters)

		pureSgInitAppUrl := os.Getenv("PURE_SG_DOMAIN") + "/ajax/public/ios/initial-app"
		fmt.Println(pureSgInitAppUrl)
		req, _ := http.NewRequest("POST", pureSgInitAppUrl, bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Println("Error creating request:", err)
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Secret-Data", "true")
		req.Header.Set("X-Forwarded-For", c.ClientIP())

		util.Log().Info("HEADERR")
		fmt.Println("X-Forwarded-For : ", c.ClientIP())

		client := &http.Client{}
		resp, respErr := client.Do(req)
		if respErr != nil {
			fmt.Println("Error sending request:", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		fmt.Println("body")
		fmt.Println(string(body))
		var respData CallbackSecretData
		_ = json.Unmarshal(body, &respData)
		respJson, _ := util.DecryptPureSGSecret(respData.SecretData)

		_ = json.Unmarshal([]byte(respJson), &twResp)

	}

	// TEMPORARY - get ab screen status from app config
	// Pending API from Taiwan
	isA := false
	abScreenString, _ := model.GetAppConfigWithCache("mode", "is_a")
	if abScreenString != "" && abScreenString == "true" {
		isA = true
	}

	// retrieve all active app domains, shuffle and return
	code = 200
	res = serializer.Response{
		Msg: "success",
		Data: DomainConfigRes{
			Api:         FindRandomDomain(model.SupportTypeApp, model.DomainTypeApi, c),
			Record:      FindRandomDomain(model.SupportTypeApp, model.DomainTypeRecord, c),
			Nami:        FindRandomDomain(model.SupportTypeApp, model.DomainTypeNami, c),
			BataceApi:   FindRandomDomain(model.SupportTypeApp, model.DomainTypeBatace, c),
			CrickongApi: FindRandomDomain(model.SupportTypeApp, model.DomainTypeCrickong, c),
			// Mode = true = Is A Screen
			// Mode = false = Is Not A Screen
			Mode:       isA,
			TaiwanMode: twResp,
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

func generateParametersByHeaderIsAB(c *gin.Context) (res map[string]interface{}) {

	// Headers
	// 设备号 - DeviceUuid
	// 是否为Ipad - IsIpad
	// VPN状态 - IsVpn
	// 充电状态 - IsCharging
	// 产品代号 - ProductCode
	// 版本 - Version

	// Headers to decide A/B screen
	// c.GetHeader("DeviceUuid")
	// c.GetHeader("IsIpad")
	// c.GetHeader("IsVpn")
	// c.GetHeader("IsCharging")
	// c.GetHeader("ProductCode")
	// c.GetHeader("Version")

	if c.GetHeader("DeviceUuid") == "" || c.GetHeader("IsIpad") == "" || c.GetHeader("IsVpn") == "" || c.GetHeader("IsCharging") == "" || c.GetHeader("ProductCode") == "" || c.GetHeader("Version") == "" {
		return
	}

	queryParameters := map[string]interface{}{
		"deviceNumber": c.GetHeader("DeviceUuid"),
		"isCharging":   false,
		"isVpn":        false,
		"isIpad":       false,
		"productCode":  c.GetHeader("ProductCode"),
		"version":      c.GetHeader("Version"),
	}

	if c.GetHeader("IsCharging") == "true" {
		queryParameters["isCharging"] = true
	}

	if c.GetHeader("IsVpn") == "true" {
		queryParameters["isVpn"] = true
	}

	if c.GetHeader("IsIpad") == "true" {
		queryParameters["isIpad"] = true
	}

	jsonData, err := json.Marshal(queryParameters)
	if err != nil {
		log.Fatalf("Error encoding JSON: %v", err)
	}

	// Print the JSON result as a string
	fmt.Println(string(jsonData))
	abRes, _ := util.EncryptPureSGSecret(string(jsonData))
	queryParameters["SecretData"] = abRes

	res = queryParameters

	return
}
