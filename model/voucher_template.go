package model

import (
	"context"

	models "blgit.rfdev.tech/taya/ploutos-object"
)

func GetPromotionVoucherTemplateByPromotionId(c context.Context, promotionID int64) (ret models.VoucherTemplate, err error) {
	err = DB.Debug().WithContext(c).Where("promotion_id", promotionID).Where("is_active").First(&ret).Error
	return
}
