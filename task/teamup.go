package task

import (
	"fmt"
	"log"
	"time"
	"web-api/model"

	models "blgit.rfdev.tech/taya/ploutos-object"
	"gorm.io/gorm"
)

func UpdateTeamupToFailIfIncomplete() {
	log.Printf("Update Teamup To Fail If Incomplete Timestamp - %s\n", fmt.Sprint(time.Now().UTC().Unix()))
	err := model.DB.Transaction(func(tx *gorm.DB) (err error) {
		err = tx.Model(&models.Teamup{}).
			Where("teamup_end_time < ?", time.Now().UTC().Unix()).
			Where("status = 0").
			Update("status", models.TeamupStatusFail).Error

		if err != nil {
			log.Printf("Fail to update to fail 1, %s\n", err.Error())
		}
		return
	})
	if err != nil {
		log.Printf("Fail to update to fail 2, %s\n", err.Error())
	}
}
