package dollar_jackpot

import (
	"context"
	"fmt"
	"strconv"
	"time"
	"web-api/cache"
	"web-api/model"
	"web-api/serializer"
	"web-api/service/common"
	"web-api/util/i18n"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

const (
	DollarJackpotRedisKey = "dollar_jackpot:%d"
)

type DollarJackpotGetService struct {
	StreamerId int `form:"streamer_id" json:"streamer_id"`
}

func (service *DollarJackpotGetService) Get(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	var dollarJackpotDraw model.DollarJackpotDraw
	var draws []model.DollarJackpotDraw
	brand := c.MustGet(`_brand`).(int)
	cacheInfo := model.CacheInfo{
		Prefix: fmt.Sprintf(`query:dollar_jackpot:%d:`, brand),
		Ttl:    10,
	}
	ctx := context.WithValue(context.TODO(), model.KeyCacheInfo, cacheInfo)
	err = model.DB.WithContext(ctx).Joins(`JOIN dollar_jackpots ON dollar_jackpots.status = 1 AND dollar_jackpots.id = dollar_jackpot_draws.dollar_jackpot_id AND dollar_jackpots.brand_id = ?`, brand).
		Where(`dollar_jackpot_draws.status`, 0).Where(`dollar_jackpots.streamer_id`, service.StreamerId).Order(`start_time`).Preload(`DollarJackpot`).Find(&draws).Error
	if err != nil {
		r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}
	for _, d := range draws {
		if d.ID != 0 && d.DollarJackpot != nil {
			if time.Now().After(d.StartTime) && time.Now().Before(d.EndTime) {
				dollarJackpotDraw = d
			}
		}
	}
	var data *serializer.DollarJackpotDraw
	if dollarJackpotDraw.ID == 0 { // if there is no ongoing draw
		var djd model.DollarJackpotDraw
		err = model.DB.Debug().WithContext(ctx).Joins("Winner").Joins(`JOIN dollar_jackpots ON dollar_jackpots.status = 1 AND dollar_jackpots.id = dollar_jackpot_draws.dollar_jackpot_id AND dollar_jackpots.brand_id = ?`, brand).
			Where(`dollar_jackpot_draws.status != 0`).Where(`dollar_jackpots.streamer_id`, service.StreamerId).Order(`start_time DESC`).Preload(`DollarJackpot`).Limit(1).Find(&djd).Error
		if err != nil {
			r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
			return
		}
		if djd.ID != 0 {
			dollarJackpotDraw = djd
			dollarJackpotDraw.Total = &dollarJackpotDraw.DollarJackpot.Prize
		}
	}
	data, err = prepareObj(c, dollarJackpotDraw)
	if err != nil {
		r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}
	r = serializer.Response{
		Data: data,
	}
	return
}

func prepareObj(c *gin.Context, dollarJackpotDraw model.DollarJackpotDraw) (data *serializer.DollarJackpotDraw, err error) {
	if dollarJackpotDraw.ID == 0 {
		return 
	}
	if dollarJackpotDraw.Total == nil {
		res := cache.RedisClient.Get(context.TODO(), fmt.Sprintf(DollarJackpotRedisKey, dollarJackpotDraw.ID))
		if res.Err() != nil && res.Err() != redis.Nil {
			err = res.Err()
			return
		}
		var total int
		if res.Val() != "" {
			total, err = strconv.Atoi(res.Val())
		}
		if err != nil {
			return
		}
		tt := int64(total)
		dollarJackpotDraw.Total = &tt
	}

	var contribution *int64
	u, isUser := c.Get("user")
	if isUser {
		user := u.(model.User)
		var sum model.ContributionSum
		err = model.DB.Model(ploutos.DollarJackpotBetReport{}).Scopes(model.GetContribution(user.ID, dollarJackpotDraw.ID)).Find(&sum).Error
		if err != nil {
			return
		}
		contribution = &sum.Sum
	}

	t := serializer.BuildDollarJackpotDraw(c, dollarJackpotDraw, contribution)
	data = &t
	return
}

type DollarJackpotWinnersService struct {
	common.Page
	JackPotId int `json:"jackpot_id" form:"jackpot_id"`
}

func (service *DollarJackpotWinnersService) List(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	brand := c.MustGet(`_brand`).(int)
	var dollarJackpotDraws []model.DollarJackpotDraw
	err = model.DB.Model(model.DollarJackpotDraw{}).Scopes(model.Paginate(service.Page.Page, service.Page.Limit)).Preload(`Winner`).
		InnerJoins(`DollarJackpot`).Order(`start_time DESC`).
		Where(`winner_id != ?`, 0).Where(`DollarJackpot.brand_id`, brand).Where(`dollar_jackpot_id`, service.JackPotId).Find(&dollarJackpotDraws).Error
	if err != nil {
		r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}

	var data []serializer.DollarJackpotDraw
	for _, d := range dollarJackpotDraws {
		data = append(data, serializer.BuildDollarJackpotDraw(c, d, nil))
	}

	r = serializer.Response{
		Data: data,
	}
	return
}


type DollarJackpotBetReportService struct {
	common.Page
	JackPotId int `json:"jackpot_id" form:"jackpot_id"`
}

func (service *DollarJackpotBetReportService) List(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	u,_:=c.Get("user")
	user := u.(model.User)
	var dollarJackpotBetReport []model.DollarJackpotBetReport
	err = model.DB.Model(model.DollarJackpotBetReport{}).Scopes(model.Paginate(service.Page.Page, service.Page.Limit)).Joins("JackpotDraws").Joins("JackpotDraws.DollarJackpot").Where(`JackpotDraws.dollar_jackpot_id`, service.JackPotId).Order(`created_at DESC`).
		Where(`user_id = ?`, user.ID).Find(&dollarJackpotBetReport).Error
	if err != nil {
		r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}
	
	var data []serializer.DollarJackpotBetReportResponse
	for _, d := range dollarJackpotBetReport {
		data = append(data, serializer.BuildDollarJackpotBetReportResponse(c, d, nil))
	}

	r = serializer.Response{
		Data: data,
	}
	return
}