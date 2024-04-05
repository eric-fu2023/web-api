package referral

import (
	"github.com/gin-gonic/gin"
	"web-api/api"
	"web-api/service/referral"
)

func GetRewardSummary(c *gin.Context) {
	var service referral.RewardSummaryService
	if err := c.ShouldBind(&service); err == nil {
		res, e := service.Get(c)
		c.JSON(200, res)
		if e != nil {
			c.Abort()
		}
	} else {
		c.JSON(400, api.ErrorResponse(c, service, err))
	}
}

func ListRewardReferrals(c *gin.Context) {
	var service referral.RewardReferralsService
	if err := c.ShouldBind(&service); err == nil {
		res, e := service.List(c)
		c.JSON(200, res)
		if e != nil {
			c.Abort()
		}
	} else {
		c.JSON(400, api.ErrorResponse(c, service, err))
	}
}

func GetRewardReferralSummary(c *gin.Context) {
	var service referral.RewardReferralSummaryService
	if err := c.ShouldBind(&service); err == nil {
		res, e := service.Get(c)
		c.JSON(200, res)
		if e != nil {
			c.Abort()
		}
	} else {
		c.JSON(400, api.ErrorResponse(c, service, err))
	}
}

func GetRewardReferralRewardRecords(c *gin.Context) {
	var service referral.RewardReferralRewardRecordsService
	if err := c.ShouldBind(&service); err == nil {
		res, e := service.List(c)
		c.JSON(200, res)
		if e != nil {
			c.Abort()
		}
	} else {
		c.JSON(400, api.ErrorResponse(c, service, err))
	}
}
