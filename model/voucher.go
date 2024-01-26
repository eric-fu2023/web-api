package model

import models "blgit.rfdev.tech/taya/ploutos-object"

type Voucher struct {
	models.Voucher
}

func VoucherGetByUserSession(userID int64, promotionSessionID int64) (v Voucher, err error) {
	err = DB.Where("user_id", userID).Where("promotion_session_id", promotionSessionID).Order("created_at desc").First(&v).Error
	return
}
