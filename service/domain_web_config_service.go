package service

import (
	"strings"
	"time"
	"web-api/model"
	"web-api/serializer"

	"github.com/gin-gonic/gin"
)

type DomainWebConfigService struct {
}

type DomainWebConfigRes struct {
	Redirect string `json:"re"`
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
}

func (service *DomainWebConfigService) InitRoute(c *gin.Context) (res serializer.Response, err error) {
	// TODO: check if the request comes from countries that should be blocked
	res = serializer.Response{
		Data: DomainWebConfigRes{
			Redirect: retrieveRandomRedirect(c),
		},
	}
	return
}

func (service *DomainConfigService) InitWeb(c *gin.Context) (res serializer.Response, err error) {
	// TODO: check if the request comes from countries that should be blocked
	// 1. find redirect
	if redirect := retrieveRandomRedirect(c); len(redirect) > 0 {
		res = serializer.Response{
			Data: DomainWebConfigRes{
				Redirect: redirect,
			},
		}
		return
	}
	// 2. if no redirects, find random domains for api/logging/nami
	res = serializer.Response{
		Data: DomainConfigRes{
			Api:         FindRandomDomain(model.SupportTypeWeb, model.DomainTypeApi, c),
			Record:      FindRandomDomain(model.SupportTypeWeb, model.DomainTypeRecord, c),
			Nami:        FindRandomDomain(model.SupportTypeWeb, model.DomainTypeNami, c),
			BataceApi:   FindRandomDomain(model.SupportTypeWeb, model.DomainTypeBatace, c),
			CrickongApi: FindRandomDomain(model.SupportTypeWeb, model.DomainTypeCrickong, c),
		},
	}
	return
}

func (service *DomainWebConfigService) RetrieveChannel(c *gin.Context) string {
	if domain := findDomainWebConfig(c); domain != nil && domain.ID > 0 {
		return domain.Channel
	}
	return ""
}

func retrieveRandomRedirect(c *gin.Context) string {
	if domain := findDomainWebConfig(c); domain != nil && domain.ID > 0 && domain.Redirect {
		redirectTos := domain.RedirectTos
		if size := len(redirectTos); size > 0 {
			return redirectTos[time.Now().UTC().UnixMicro()%int64(size)].Origin
		} else {
			return ""
		}
	}
	return ""
}

func findDomainWebConfig(c *gin.Context) (domain *model.DomainWebConfig) {
	if origin := c.Request.Header.Get("ori"); len(strings.TrimSpace(origin)) > 0 {
		domain = model.DomainWebConfig{}.FindByOrigin(c, origin)
	}
	return
}
