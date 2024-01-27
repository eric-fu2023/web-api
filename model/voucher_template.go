package model

import (
	"context"

	models "blgit.rfdev.tech/taya/ploutos-object"
)

func VoucherTemplateGetByPromotion(c context.Context, promotionID int64) (ret models.VoucherTemplate, err error) {
	err = DB.WithContext(c).Where("promotion_id", promotionID).First(&ret).Error
	return
}
