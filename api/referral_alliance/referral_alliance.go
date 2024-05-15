package referral_alliance

import (
	"github.com/gin-gonic/gin"
	"web-api/api"
	"web-api/service/referral_alliance"
)

func GetSummary(c *gin.Context) {
	var service referral_alliance.SummaryService
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

func ListReferrals(c *gin.Context) {
	var service referral_alliance.ReferralsService
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

func GetReferralSummary(c *gin.Context) {
	var service referral_alliance.ReferralSummaryService
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

func GetReferralRewardRecords(c *gin.Context) {
	var service referral_alliance.ReferralRewardRecordsService
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

func GetRankings(c *gin.Context) {
	var service referral_alliance.RankingsService
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
