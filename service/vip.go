package service

import (
	"web-api/model"
	"web-api/serializer"
	"web-api/util"

	"github.com/gin-gonic/gin"
)

const vip_rules_full = "vip_rules_full"

type VipQuery struct {
}

func (s VipQuery) Get(c *gin.Context) serializer.Response {
	user := c.MustGet("user").(model.User)

	vip, err := model.GetVipWithDefault(c, user.ID)
	if err != nil {
		return serializer.Err(c, s, serializer.CodeGeneralError, "", err)
	}
	return serializer.Response{
		Data: serializer.BuildVip(vip),
	}
}

type VipLoad struct {
}

func (s VipLoad) Load(c *gin.Context) serializer.Response {

	list, err := model.LoadVipRule(c)
	if err != nil {
		return serializer.Err(c, s, serializer.CodeGeneralError, "", err)
	}

	return serializer.Response{
		Data: util.MapSlice(list, serializer.BuildVipRule),
	}
}

// func CachedVipRules(ctx context.Context) ([]models.VIPRule, error) {
// 	res := cache.RedisClient.Get(ctx, vip_rules_full)
// 	if res.Err() != nil {
// 		list, err := model.LoadRule(ctx)
// 		go cache.RedisClient.Set(ctx, vip_rules_full, list, 5*time.Minute)
// 		return list, err
// 	}
// 	ret := []models.VIPRule{}
// 	err := json.Unmarshal([]byte(res.Val()), &ret)
// 	return ret, err
// }
