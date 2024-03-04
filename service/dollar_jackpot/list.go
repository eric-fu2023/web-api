package dollar_jackpot

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"strconv"
	"time"
	"web-api/cache"
	"web-api/model"
	"web-api/serializer"
	"web-api/service/common"
	"web-api/util/i18n"
)

const (
	DollarJackpotRedisKey = "dollar_jackpot:%d"
)

type DollarJackpotGetService struct {
}

func (service *DollarJackpotGetService) Get(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	var dollarJackpotDraw model.DollarJackpotDraw
	brand := c.MustGet(`_brand`).(int)
	cacheInfo := model.CacheInfo{
		Prefix: fmt.Sprintf(`query:dollar_jackpot:%d:`, brand),
		Ttl:    10,
	}
	ctx := context.WithValue(context.TODO(), model.KeyCacheInfo, cacheInfo)
	err = model.DB.WithContext(ctx).Joins(`JOIN dollar_jackpots ON dollar_jackpots.id = dollar_jackpot_draws.dollar_jackpot_id AND dollar_jackpots.brand_id = ?`, brand).
		Where(`dollar_jackpot_draws.status`, 0).Order(`start_time`).
		Preload(`DollarJackpot`).Limit(1).Find(&dollarJackpotDraw).Error
	if err != nil {
		r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}
	var data *serializer.DollarJackpotDraw
	if dollarJackpotDraw.ID != 0 && dollarJackpotDraw.DollarJackpot != nil && time.Now().Before(dollarJackpotDraw.EndTime) && time.Now().After(dollarJackpotDraw.StartTime) {
		res := cache.RedisClient.Get(context.TODO(), fmt.Sprintf(DollarJackpotRedisKey, dollarJackpotDraw.ID))
		if res.Err() != nil && res.Err() != redis.Nil {
			r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), res.Err())
			return
		}
		var total int
		if res.Val() != "" {
			total, err = strconv.Atoi(res.Val())
		}
		if err != nil {
			r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
			return
		}
		tt := int64(total)
		dollarJackpotDraw.Total = &tt

		t := serializer.BuildDollarJackpotDraw(c, dollarJackpotDraw)
		data = &t
	}
	r = serializer.Response{
		Data: data,
	}
	return
}

type DollarJackpotWinnersService struct {
	common.Page
}

func (service *DollarJackpotWinnersService) List(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	brand := c.MustGet(`_brand`).(int)
	var dollarJackpotDraws []model.DollarJackpotDraw
	err = model.DB.Model(model.DollarJackpotDraw{}).Scopes(model.Paginate(service.Page.Page, service.Page.Limit)).Preload(`Winner`).
		InnerJoins(`DollarJackpot`).Order(`start_time DESC`).
		Where(`winner_id != ?`, 0).Where(`DollarJackpot.brand_id`, brand).Find(&dollarJackpotDraws).Error
	if err != nil {
		r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}

	var data []serializer.DollarJackpotDraw
	for _, d := range dollarJackpotDraws {
		data = append(data, serializer.BuildDollarJackpotDraw(c, d))
	}

	r = serializer.Response{
		Data: data,
	}
	return
}
