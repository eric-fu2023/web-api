package model

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"gorm.io/gorm"
)

type Teamup struct {
	ploutos.Teamup
}

func SaveTeamup(teamup ploutos.Teamup) (err error) {
	err = DB.Transaction(func(tx *gorm.DB) (err error) {
		err = tx.Save(&teamup).Error
		return
	})

	return
}

func GetTeamUp(key string, value string) (teamup ploutos.Teamup, err error) {

	err = DB.Transaction(func(tx *gorm.DB) (err error) {
		err = tx.Where("? = ?", key, value).First(&teamup).Error
		return
	})

	return
}

func GetAllTeamUps(userId int64, status []int, page, limit int) (teamups []ploutos.Teamup, err error) {

	err = DB.Transaction(func(tx *gorm.DB) (err error) {
		tx = tx.Where("user_id = ?", userId)

		if len(status) > 0 {
			tx = tx.Where("status in ?", status)
		}

		err = tx.Scopes(Paginate(page, limit)).Find(&teamups).Error

		return
	})

	return
}
