package model

import (
	"database/sql"
	"errors"
	"time"
	"web-api/conf/consts"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var (
	ErrAccountLimitExceeded = errors.New("withdraw_account_limit_exceeded")
)

type UserAccountBinding struct {
	ploutos.UserAccountBinding
	CashMethod *CashMethod
}

func (UserAccountBinding) GetAccountByUser(c *gin.Context, userID, vipID int64) (list []UserAccountBinding, err error) {
	user := c.MustGet("user").(User)

	q := DB.Joins("CashMethod").
		Where("user_account_binding.user_id", userID).
		Where("user_account_binding.is_active").
		Order("user_account_binding.id desc")

	var restrictPaymentChannel []int64 = user.RestrictPaymentChannel
	if len(restrictPaymentChannel) != 0 {
		q = q.Where("\"CashMethod\".id NOT IN ?", restrictPaymentChannel)
	}

	now := time.Now().UTC()
	q = q.Joins("CashMethod.CashMethodPromotion", DB.
		Where("\"CashMethod__CashMethodPromotion\".start_at < ? and \"CashMethod__CashMethodPromotion\".end_at > ?", now, now).
		Where("\"CashMethod__CashMethodPromotion\".status = ?", 1).
		Where("\"CashMethod__CashMethodPromotion\".vip_id = ?", vipID),
	)

	err = q.Find(&list).Error
	return
}

func (b *UserAccountBinding) AddToDb() (err error) {
	err = DB.Transaction(func(tx *gorm.DB) (err error) {
		var num int64
		err = tx.Model(&UserAccountBinding{}).Where("user_id", b.UserID).Where("is_active").Count(&num).Error
		if err != nil {
			return
		}
		if num >= consts.WithdrawMethodLimit {
			return ErrAccountLimitExceeded
		}
		err = tx.Create(b).Error
		return
	}, &sql.TxOptions{Isolation: sql.LevelSerializable})
	return
}

func (b *UserAccountBinding) SoftRemove() (err error) {
	err = DB.Model(&UserAccountBinding{}).Where(&b).Update("is_active", false).Error
	return
}

func (b *UserAccountBinding) HardRemove() error {
	return DB.Model(&UserAccountBinding{}).Delete(&b).Where("id", b.ID).Error
}
