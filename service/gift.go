package service

import (
	"web-api/serializer"

	"web-api/model"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"github.com/gin-gonic/gin"
)

type GiftListService struct {
}

type GiftListResponse struct {
}

func (service *GiftListService) List(c *gin.Context) (r serializer.Response, err error) {
	var gifts []ploutos.Gift
	// i18n := c.MustGet("i18n").(i18n.I18n)

	err = model.DB.Model(ploutos.Gift{}).Scopes(model.ByActiveGifts(true)).Find(&gifts).Error
	if err != nil {
		return
	}

	r = serializer.Response{
		Data: serializer.BuildGift(gifts),
	}
	return
}
