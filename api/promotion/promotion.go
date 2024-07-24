package promotion

import (
	"web-api/api"
	"web-api/service/promotion"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func GetCategoryList(c *gin.Context) {
	var service promotion.PromotionList
	if err := c.ShouldBind(&service); err == nil {
		res, _ := service.ListCategories(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, api.ErrorResponse(c, service, err))
	}
}

func GetCoverList(c *gin.Context) {
	var service promotion.PromotionList
	if err := c.ShouldBind(&service); err == nil {
		res, _ := service.Handle(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, api.ErrorResponse(c, service, err))
	}
}

func GetDetail(c *gin.Context) {
	var service promotion.PromotionDetail
	if err := c.ShouldBind(&service); err == nil {
		res, _ := service.Handle(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, api.ErrorResponse(c, service, err))
	}
}

func GetCustomDetail(c *gin.Context) {
	var service promotion.PromotionCustomDetail
	if err := c.ShouldBind(&service); err == nil {
		res, _ := service.Handle(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, api.ErrorResponse(c, service, err))
	}
}

func PromotionClaim(c *gin.Context) {
	var service promotion.PromotionClaim
	if err := c.ShouldBind(&service); err == nil {
		res, _ := service.Handle(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, api.ErrorResponse(c, service, err))
	}
}

func PromotionJoin(c *gin.Context) {
	var service promotion.PromotionJoin
	if err := c.ShouldBindWith(&service, binding.Form); err == nil {
		res, _ := service.Handle(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, api.ErrorResponse(c, service, err))
	}
}

func VoucherList(c *gin.Context) {
	var service promotion.VoucherList
	if err := c.ShouldBind(&service); err == nil {
		res, _ := service.Handle(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, api.ErrorResponse(c, service, err))
	}
}

func VoucherApplicable(c *gin.Context) {
	var service promotion.VoucherApplicable
	if err := c.ShouldBindJSON(&service); err == nil {
		res, _ := service.Handle(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, api.ErrorResponse(c, service, err))
	}
}

func VoucherPreBinding(c *gin.Context) {
	var service promotion.VoucherPreBinding
	if err := c.ShouldBindJSON(&service); err == nil {
		res, _ := service.Handle(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, api.ErrorResponse(c, service, err))
	}
}
