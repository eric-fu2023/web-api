package model

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"web-api/util"
	"web-api/util/i18n"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"gorm.io/gorm"
)

const (
	InitialRandomFakeProgressLowerLimit = int64(8500)
	InitialRandomFakeProgressUpperLimit = int64(9200)

	// Production Rate
	SubsequentRandomFakeProgressLowerLimit = int64(1)
	SubsequentRandomFakeProgressUpperLimit = int64(100)

	// SubsequentRandomFakeProgressLowerLimit = int64(400)
	// SubsequentRandomFakeProgressUpperLimit = int64(700)
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
			Where("teamup_entries.teamup_id = ?", teamupId).
			Order(`teamup_entries.created_at DESC`)

		tx = tx.Scopes(Paginate(page, limit))

		if err := tx.Scan(&res).Error; err != nil {
			return err
		}
		return nil
	})

	return
}

func CreateSlashBetRecord(teamupId, userId int64, i18n i18n.I18n) (teamup ploutos.Teamup, isTeamupSuccess, isSuccess bool, err error) {

	// First entry - 85% ~ 92%
	// Second entry onwards until N - 1 - 0.01% ~ 1%
	// Capped at 99.99% if deposit not fulfilled

	// LAST SLASH - if deposit fulfilled then slash - will make it 100% and complete it

	// NO SLASH if user slashed before

	teamup, _ = GetTeamUpByTeamUpId(teamupId)

	if teamup.UserId == userId || teamup.TeamupEndTime < time.Now().UTC().Unix() {
		err = fmt.Errorf("teamup_slash_error")
		return
	}

	teamupEntries, err := FindTeamupEntryByTeamupId(teamupId)
	if err != nil {
		return
	}

	for _, entry := range teamupEntries {
		if entry.UserId == userId {
			err = fmt.Errorf("teamup_slashed_before_error")
			return
		}
	}

	// 10000 = 100%
	maxPercentage := int64(10000)
	currentTotalProgress := util.Sum(teamupEntries, func(entry ploutos.TeamupEntry) int64 {
		return entry.FakePercentageProgress
	})

	beforeProgress, afterProgress := GenerateFakeProgress(currentTotalProgress)

	// Update status to SUCCESS if teamup deposit exceeded target
	if teamup.TotalTeamupDeposit >= teamup.TotalTeamUpTarget {
		isTeamupSuccess = true
		teamup.Status = int(ploutos.TeamupStatusSuccess)
		teamup.TeamupCompletedTime = time.Now().UTC().Unix()
		afterProgress = maxPercentage
	}

	slashEntry := ploutos.TeamupEntry{
		TeamupId: teamupId,
		UserId:   userId,
	}

	slashEntry.TeamupEndTime = teamup.TeamupEndTime
	slashEntry.TeamupCompletedTime = teamup.TeamupCompletedTime
	slashEntry.FakePercentageProgress = afterProgress - beforeProgress

	teamup.TotalFakeProgress = afterProgress

	err = DB.Transaction(func(tx *gorm.DB) (err error) {
		err = tx.Save(&teamup).Error
		return
	})
	if err != nil {
		return
	}

	err = DB.Transaction(func(tx *gorm.DB) (err error) {
		err = tx.Save(&slashEntry).Error
		return
	})

	isSuccess = true

	return
}

func FindOngoingTeamupEntriesByUserId(userId int64) (res ploutos.TeamupEntry, err error) {
	err = DB.Transaction(func(tx *gorm.DB) error {
		tx = tx.Table("teamup_entries").
			Select("teamup_entries.*"). // Select fields from the teamup_entries table
			Joins("JOIN teamups ON teamups.id = teamup_entries.teamup_id").
			Where("teamup_entries.user_id = ?", userId).
			Where("teamups.status = 0").
			Order("teamup_entries.created_at ASC").
			First(&res) // Fetch the first matching record

		if err := tx.Error; err != nil {
			return err
		}

		return nil
	})

	return
}

func UpdateFirstTeamupEntryProgress(tx *gorm.DB, teamupEntryId, amount, slashAmount int64) error {
	return tx.Transaction(func(tx2 *gorm.DB) error {

		updates := map[string]interface{}{
			"total_deposit":              gorm.Expr("total_deposit + ?", amount),
			"contributed_teamup_deposit": gorm.Expr("contributed_teamup_deposit + ?", slashAmount),
		}

		if err := tx2.Table("teamup_entries").
			Where("id = ?", teamupEntryId).
			Limit(1).
			Updates(updates).Error; err != nil {
			return err
		}
		return nil
	})
}

func GenerateFakeProgress(currentProgress int64) (beforeProgress, afterProgress int64) {
	ceilingProgress := int64(9999)
	beforeProgress = currentProgress

	if currentProgress >= ceilingProgress {
		beforeProgress = ceilingProgress
		afterProgress = ceilingProgress
		return
	}

	maximumAllowedProgress := ceilingProgress - currentProgress

	if currentProgress == 0 {
		afterProgress = util.RandomNumFromRange(InitialRandomFakeProgressLowerLimit, InitialRandomFakeProgressUpperLimit)
		return
	}

	lowerLimit, err := GetAppConfigWithCache("teamup", "teamup_subsequent_fake_progress_lower")
	if err != nil {
		lowerLimit = "1"
	}
	upperLimit, err := GetAppConfigWithCache("teamup", "teamup_subsequent_fake_progress_upper")
	if err != nil {
		upperLimit = "100"
	}
	subsequentLowerLimit, err := strconv.Atoi(lowerLimit)
	if err != nil {
		subsequentLowerLimit = 1
	}
	subsequentUpperLimit, err := strconv.Atoi(upperLimit)
	if err != nil {
		subsequentUpperLimit = 100
	}

	progress := int64(math.Min(float64(maximumAllowedProgress), float64(util.RandomNumFromRange(int64(subsequentLowerLimit), int64(subsequentUpperLimit)))))
	afterProgress = progress + currentProgress

	return
}

func GetTeamupEntryByTeamupIdAndUserId(teamupId, userId int64) (res ploutos.TeamupEntry, err error) {

	err = DB.Transaction(func(tx *gorm.DB) error {
		tx = tx.Table("teamup_entries").
			Where("teamup_entries.teamup_id = ?", teamupId).
			Where("teamup_entries.user_id = ?", userId).
			First(&res)

		if err := tx.Error; err != nil {
			return err
		}

		return nil
	})

	return
}
