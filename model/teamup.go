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

func GetTeamUp(orderId string) (teamup ploutos.Teamup, err error) {

	err = DB.Transaction(func(tx *gorm.DB) (err error) {
		err = tx.Model(ploutos.Teamup{}).Where("order_id = ?", orderId).First(&teamup).Error
		return
	})

	return
}

func GetAllTeamUps(userId int64, status []int, page, limit int) (teamups []ploutos.Teamup, err error) {
	var results []struct {
		UserId      string
		BetReportId string
	}
	err = DB.Transaction(func(tx *gorm.DB) (err error) {
		tx = tx.Select("teamup.user_id as user_id, bet_report.id as bet_report_id").Joins("left join bet_report on teamup.order_id = bet_report.business_id").Scan(&results)
		err = tx.Error
		// tx = tx.Where("user_id = ?", userId)

		// if len(status) > 0 {
		// 	tx = tx.Where("status in ?", status)
		// }

		// err = tx.Scopes(Paginate(page, limit)).Find(&teamups).Error

		return
	})

	return
}
