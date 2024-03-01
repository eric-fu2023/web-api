package dollar_jackpot

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"strconv"
	"time"
	"web-api/cache"
	"web-api/model"
	"web-api/serializer"
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
		Prefix: fmt.Sprintf(`query:dollar_jackpot:%d`, brand),
		Ttl:    3,
	}
	ctx := context.WithValue(context.TODO(), model.KeyCacheInfo, cacheInfo)
	err = model.DB.WithContext(ctx).Where(`winner_id`, 0).Order(`start_time`).Preload(`DollarJackpot`, func(db *gorm.DB) *gorm.DB {
		return db.Scopes(model.ByBrand(int64(brand)))
	}).Limit(1).Find(&dollarJackpotDraw).Error
	if err != nil {
		r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}
	var data *serializer.DollarJackpotDraw
	if dollarJackpotDraw.ID != 0 && dollarJackpotDraw.DollarJackpot.ID != 0 && time.Now().Before(dollarJackpotDraw.EndTime) && time.Now().After(dollarJackpotDraw.StartTime) {
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
		dollarJackpotDraw.Total = int64(total)

		t := serializer.BuildDollarJackpotDraw(dollarJackpotDraw)
		data = &t
	}
	r = serializer.Response{
		Data: data,
	}
	return
}
