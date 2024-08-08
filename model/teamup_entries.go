package model

import (
	"time"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"gorm.io/gorm"
)

type TeamupEntry struct {
	ploutos.TeamupEntry
}

type TeamupEntryCustomRes []struct {
	ContributedAmount float64   `json:"contributed_amount"`
	ContributedTime   time.Time `json:"contributed_time"`
	Nickname          string    `json:"nickname"`
	Avatar            string    `json:"avatar"`
}

func SaveTeamupEntry(teamupEntry ploutos.TeamupEntry) (err error) {
	err = DB.Transaction(func(tx *gorm.DB) (err error) {
		err = tx.Save(&teamupEntry).Error
		return
	})

	return
}
func GetAllTeamUpEntries(teamupId int64, page, limit int) (res TeamupEntryCustomRes, err error) {

	err = DB.Transaction(func(tx *gorm.DB) error {
		tx = tx.Table("teamup_entries").
			Select("teamup_entries.contributed_teamup_deposit as contributed_amount, teamup_entries.created_at as contributed_time, users.nickname as nickname, users.avatar as avatar").
			Joins("left join users on teamup_entries.user_id = users.id").
			Where("teamup_entries.teamup_id = ?", teamupId)

		tx = tx.Scopes(Paginate(page, limit))

		if err := tx.Scan(&res).Error; err != nil {
			return err
		}
		return nil
	})

	return
}
