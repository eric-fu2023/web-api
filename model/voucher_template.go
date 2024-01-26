package model

import (
	"context"

	models "blgit.rfdev.tech/taya/ploutos-object"
)

type VoucherTemplate struct {
	models.VoucherTemplate
}

func VoucherTemplateGetByPromotion(c context.Context, promotionID int64) (ret VoucherTemplate, err error) {
	err = DB.WithContext(c).Where("id", promotionID).First(&ret).Error
	return
}
