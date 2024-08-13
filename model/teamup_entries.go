package model

import (
	"time"

	"web-api/util"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"gorm.io/gorm"
)

type TeamupEntryCustomRes []struct {
	ContributedAmount float64   `json:"contributed_amount"`
	ContributedTime   time.Time `json:"contributed_time"`
	Nickname          string    `json:"nickname"`
	Avatar            string    `json:"avatar"`
	Progress          int64     `json:"progress"`
}

func FindTeamupEntryByTeamupId(teamupId int64) (res []ploutos.TeamupEntry, err error) {
	err = DB.Transaction(func(tx *gorm.DB) error {
		tx = tx.Table("teamup_entries").Where("teamup_entries.teamup_id = ?", teamupId).Find(&res)
		if err := tx.Scan(&res).Error; err != nil {
			return err
		}
		return nil
	})

	return
}

func GetAllTeamUpEntries(teamupId int64, page, limit int) (res TeamupEntryCustomRes, err error) {

	err = DB.Transaction(func(tx *gorm.DB) error {
		tx = tx.Table("teamup_entries").
			Select("teamup_entries.contributed_teamup_deposit as contributed_amount, teamup_entries.created_at as contributed_time, teamup_entries.fake_percentage_progress as progress, users.nickname as nickname, users.avatar as avatar").
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

func CreateSlashBetRecord(teamupId, userId int64) (isSuccess bool, err error) {

	// First entry - 85% ~ 92%
	// Second entry onwards until N - 1 - 0.01% ~ 1%
	// Capped at 99.99% if deposit not fulfilled

	// LAST SLASH - if deposit fulfilled then slash - will make it 100% and complete it

	// NO SLASH if user slashed before

	teamupEntries, err := FindTeamupEntryByTeamupId(teamupId)
	if err != nil {
		return
	}

	for _, entry := range teamupEntries {
		if entry.UserId == userId {
			return
		}
	}

	// 10000 = 100%
	maxPercentage := int64(10000)
	ceilingPercentage := int64(9999)
	totalProgress := util.Sum(teamupEntries, func(entry ploutos.TeamupEntry) int64 {
		return entry.FakePercentageProgress
	})

	var rNum int64
	var currentSlashProgress int64

	if totalProgress == 0 {
		rNum = util.RandomNumFromRange(int64(8500), int64(9200))
	} else {
		rNum = util.RandomNumFromRange(int64(1), int64(100))
	}

	totalProgress += int64(rNum)

	slashEntry := ploutos.TeamupEntry{
		TeamupId: teamupId,
		UserId:   userId,
	}

	currentSlashProgress = 0

	teamup, _ := GetTeamUpByTeamUpId(teamupId)
	isSuccessTeamup := false

	slashEntry.TeamupEndTime = teamup.TeamupEndTime
	slashEntry.TeamupCompletedTime = teamup.TeamupCompletedTime

	// Update status to SUCCESS if teamup deposit exceeded target
	if teamup.TotalTeamupDeposit >= teamup.TotalTeamUpTarget {
		isSuccessTeamup = true
		teamup.Status = int(ploutos.TeamupStatusSuccess)
		err = DB.Transaction(func(tx *gorm.DB) (err error) {
			err = tx.Save(&teamup).Error
			return
		})

		if err != nil {
			return
		}
	}

	if isSuccessTeamup {
		// If already success, add the progress until 100%
		currentSlashProgress = maxPercentage - totalProgress
		teamup.TotalFakeProgress = maxPercentage
	} else if totalProgress >= maxPercentage-1 {
		// If not success, and totalProgress more than or equals to 99.99%
		if totalProgress-rNum >= ceilingPercentage {
			// If not success, and totalProgress is 99.99%, current slash = 0.00%
			currentSlashProgress = 0
		} else {
			// If not success, and totalProgress less than 99.99%, current slash = 99.99% - currentProgress = make it 99.99%
			currentSlashProgress = ceilingPercentage - totalProgress
		}
		teamup.TotalFakeProgress = 9999
	} else {
		// If not success, normal random %
		currentSlashProgress = rNum
		teamup.TotalFakeProgress = totalProgress + rNum
	}

	err = DB.Transaction(func(tx *gorm.DB) (err error) {
		err = tx.Save(&teamup).Error
		return
	})
	if err != nil {
		return
	}

	slashEntry.FakePercentageProgress = currentSlashProgress

	err = DB.Transaction(func(tx *gorm.DB) (err error) {
		err = tx.Save(&slashEntry).Error
		return
	})

	isSuccess = true

	return
}
