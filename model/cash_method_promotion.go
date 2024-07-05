package model

import (
	models "blgit.rfdev.tech/taya/ploutos-object"
	"gorm.io/gorm"
)

func FindCashMethodPromotionByCashMethodIdAndVipId(cashMethodId, vipId int64, tx *gorm.DB) (cashMethodPromotion models.CashMethodPromotion, err error) {
	if tx == nil {
		tx = DB
	}
	err = tx.Where("cash_method_id", cashMethodId).Where("vip_id", vipId).Find(&cashMethodPromotion).Error
	if err != nil {
		return
	}
	return
}
