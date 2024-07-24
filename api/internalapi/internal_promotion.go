package internal_api

import (
	"web-api/serializer"
	internalclaim "web-api/service/promotion/internal_claim"

	"github.com/gin-gonic/gin"
)

func InternalPromotion(c *gin.Context) {
	var service internalclaim.VipBatchClaimRequest
	if err := c.ShouldBind(&service); err == nil {
		if res, err := service.Handle(c); err == nil {
			c.JSON(200, res)
		} else {
			c.JSON(500, serializer.EnsureErr(c, err, res))
		}
	} else {
		c.JSON(400, serializer.ParamErr(c, service, "", err))
	}
}

func InternalPromotionRequest(c *gin.Context) {
	var service internalclaim.CustomPromotionClaimRequest
	if err := c.ShouldBind(&service); err == nil {
		if res, err := service.Handle(c); err == nil {
			c.JSON(200, res)
		} else {
			c.JSON(500, serializer.EnsureErr(c, err, res))
		}
	} else {
		c.JSON(400, serializer.ParamErr(c, service, "", err))
	}
}
