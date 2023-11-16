package model

import (
	"database/sql"
	"errors"
	"web-api/conf/consts"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"gorm.io/gorm"
)

type UserAccountBinding struct {
	ploutos.UserAccountBinding
	CashMethod *CashMethod
}

func (UserAccountBinding) GetAccountByUser(userID int64) (list []UserAccountBinding, err error) {
	err = DB.Preload("CashMethod").Where("user_id", userID).Where("is_active").Order("id desc").Find(&list).Error
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
			return errors.New("limit exceeded")
		}
		err = tx.Create(b).Error
		return
	}, &sql.TxOptions{Isolation: sql.LevelSerializable})
	return
}

func (b *UserAccountBinding) Remove() (err error) {
	err = DB.Model(&UserAccountBinding{}).Where(&b).Update("is_active", false).Error
	return
}
