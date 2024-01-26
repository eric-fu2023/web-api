package model

import (
	"context"
	"fmt"
	"web-api/util"

	models "blgit.rfdev.tech/taya/ploutos-object"
)

type Voucher struct {
	models.Voucher
}

func AmountReplace(original string, amount int64) string {
	return util.TextReplace(original, map[string]string{
		"amount": fmt.Sprintf("%d", amount),
	})
}

func VoucherGetByUserSession(c context.Context, userID int64, promotionSessionID int64) (v Voucher, err error) {
	err = DB.WithContext(c).Where("user_id", userID).Where("promotion_session_id", promotionSessionID).Order("created_at desc").First(&v).Error
	return
}
