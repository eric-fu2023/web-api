package model

import (
	"fmt"
	"log"
	"math"
	"sort"
	"strconv"
	"time"

	"web-api/util"
	"web-api/util/i18n"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"

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

func CreateSlashBetRecord(c *gin.Context, teamupId int64, user ploutos.User, i18n i18n.I18n) (teamup ploutos.Teamup, isTeamupSuccess, isSuccess bool, err error) {

	// First entry - 85% ~ 92%
	// Second entry onwards until N - 1 - 0.01% ~ 1%
	// Capped at 99.99% if deposit not fulfilled

	// NO SLASH if user slashed before

	userId := user.ID

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

	var beforeProgress, afterProgress int64

	// 10000 = 100%
	// maxPercentage := int64(10000)
	currentTotalProgress := util.Sum(teamupEntries, func(entry ploutos.TeamupEntry) int64 {
		return entry.FakePercentageProgress
	})

	// IF THIS USER DOESNT FULFILL
	// 1) NEW USER
	// 2) TOP-UP BEFORE
	// THEN SLASH 0%

	shouldGiveRandomPercentage := false
	isValidSlash := validSlash(c, user)
	if !isValidSlash {
		shouldGiveRandomPercentage = true
	}

	// Check if teamup is shortlisted
	// If yes, then success the current shortlisted
	if teamup.Term != 0 && teamup.ShortlistStatus != ploutos.ShortlistStatusNotShortlist {
		// 如果有Term，如果这单是成功 / 入选
		if isValidSlash {
			isSuccessShortlisted, _ := SuccessShortlisted(teamup, currentTotalProgress, userId)

			// No matter got error or not, need to return
			// No error = success = return
			// Got error = should not continue = return
			if isSuccessShortlisted {
				isTeamupSuccess = isSuccessShortlisted
				return
			}

			// Give random percentage if shortlisted by others, slash for fun
			shouldGiveRandomPercentage = true
		}
	} else if teamup.Term != 0 && teamup.ShortlistStatus == ploutos.ShortlistStatusNotShortlist {

		// 如果有Term，就代表CONTRIBUTION >= TARGET+不是入选/成功，就意思意思砍0
		isValidSlash = false
		shouldGiveRandomPercentage = true
	}

	if isValidSlash {
		// 如果currentTotalProgress = 0，beforeProgress = 0，代表第一刀，afterProgress - beforeProgress的差值会比较大
		beforeProgress, afterProgress = GenerateFakeProgress(currentTotalProgress)
	} else {
		if shouldGiveRandomPercentage {
			beforeProgress, afterProgress = GenerateFakeProgress(currentTotalProgress)
		} else {
			beforeProgress = currentTotalProgress
			afterProgress = currentTotalProgress
		}
	}

	slashEntry := ploutos.TeamupEntry{
		TeamupId: teamupId,
		UserId:   userId,
	}

	slashEntry.TeamupEndTime = teamup.TeamupEndTime
	slashEntry.TeamupCompletedTime = teamup.TeamupCompletedTime
	slashEntry.FakePercentageProgress = afterProgress - beforeProgress

	teamup.TotalFakeProgress = afterProgress

	if isValidSlash {

		// 需求改动 所有砍为小砍
		teamupContributeFixedAmountString, _ := GetAppConfigWithCache("teamup", "teamup_fixed_amount")
		if teamupContributeFixedAmountString != "" {
			contributeAmount, _ := strconv.Atoi(teamupContributeFixedAmountString)
			slashEntry.ContributedTeamupDeposit = int64(contributeAmount)
			teamup.TotalTeamupDeposit += int64(contributeAmount)
		}
		// if teamup.TotalTeamupDeposit == 0 { // 如果贡献价值为0（首次），大砍
		// 	// 计算公式
		// 	// (首次砍刀百分比 / 100) * (砍单目标价值 / 100) / 100 = 砍成价值
		// 	// (6516 / 100) * (21000 / 100) / 100 = $136.836

		// 	slashValue := ((float64(slashEntry.FakePercentageProgress) / 100) * (float64(teamup.TotalTeamUpTarget) / 100)) / 100
		// 	roundedCeilSlashValue := (math.Ceil(slashValue*100) / 100) * 100

		// 	slashEntry.ContributedTeamupDeposit = int64(roundedCeilSlashValue)
		// 	teamup.TotalTeamupDeposit += int64(roundedCeilSlashValue)
		// } else { // 如果贡献价值不为0（不是首次），小砍
		// 	teamupContributeFixedAmountString, _ := GetAppConfigWithCache("teamup", "teamup_fixed_amount")
		// 	if teamupContributeFixedAmountString != "" {
		// 		contributeAmount, _ := strconv.Atoi(teamupContributeFixedAmountString)
		// 		slashEntry.ContributedTeamupDeposit = int64(contributeAmount)
		// 		teamup.TotalTeamupDeposit += int64(contributeAmount)
		// 	}
		// }
	}

	var tx *gorm.DB

	// 如果未进入候选池 + 后端数值已达标 = 加入候选池 （20进4）
	if isValidSlash && teamup.Term == 0 && teamup.TotalTeamupDeposit >= teamup.TotalTeamUpTarget {

		// afterProgress = int64(9999)
		slashEntry.FakePercentageProgress = afterProgress - beforeProgress
		teamup.TotalFakeProgress = afterProgress

		// 获得该期
		currentTerm, _ := GetCurrentTermNum()
		teamupTermSizeString, _ := GetAppConfigWithCache("teamup", "term_size")
		termSize, _ := strconv.Atoi(teamupTermSizeString)
		termTeamups, _ := FindExceedTargetByTerm(currentTerm)
		// 默认该单为这一期
		teamup.Term = currentTerm

		// 如果这期已大过一期该有的上限，就是下一期了
		if len(termTeamups)+1 > termSize {
			teamup.Term++
		}
		// 如果改单是这期最后一个
		// 选价值最小的4张单晋级，之后这4张单选一张砍单成功
		if len(termTeamups)+1 == termSize {

			// 默认为4
			teamupTermShortlistedNumString, _ := GetAppConfigWithCache("teamup", "teamup_shortlisted_num")
			termShortlistedNum, _ := strconv.Atoi(teamupTermShortlistedNumString)

			// 把这单加入并筛选出4个价值最小的
			termTeamups = append(termTeamups, teamup)
			sort.Slice(termTeamups, func(i, j int) bool {
				return int(termTeamups[i].TotalTeamUpTarget) < int(termTeamups[j].TotalTeamUpTarget)
			})
			termTeamups = termTeamups[:termShortlistedNum]
			var ids []int64
			for _, t := range termTeamups {
				if t.ID == teamup.ID {
					teamup.ShortlistStatus = ploutos.ShortlistStatusShortlisted
				}
				ids = append(ids, t.ID)
			}

			// 选价值最小的4张单晋级，之后这4张单选一张砍单成功
			err = FlagStatusShortlisted(tx, ids)
			if err != nil {
				return
			}
		}
	}

	err = tx.Transaction(func(tx2 *gorm.DB) (err error) {
		err = tx2.Save(&teamup).Error
		return
	})
	if err != nil {
		return
	}

	err = tx.Transaction(func(tx2 *gorm.DB) (err error) {
		err = tx2.Save(&slashEntry).Error
		return
	})

	if err != nil {
		tx.Rollback()
		return
	}
	tx.Commit()
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
		var configError error
		initialLowerLimitString, configError := GetAppConfigWithCache("teamup", "teamup_initial_fake_progress_lower")
		if configError != nil {
			log.Printf("Error computing initialLowerLimitString, %s\n", configError.Error())
		}
		initialUpperLimitString, configError := GetAppConfigWithCache("teamup", "teamup_initial_fake_progress_upper")
		if configError != nil {
			log.Printf("Error computing initialUpperLimitString, %s\n", configError.Error())
		}
		initialLowerLimit, _ := strconv.Atoi(initialLowerLimitString)
		initialUpperLimit, _ := strconv.Atoi(initialUpperLimitString)
		afterProgress = util.RandomNumFromRange(int64(initialLowerLimit), int64(initialUpperLimit))
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

func validSlash(c *gin.Context, user ploutos.User) (isValid bool) {

	var condition1, condition2 bool
	topUpCashOrder, _ := FirstTopup(c, user.ID)
	if topUpCashOrder.ID != "" {
		condition1 = true
	}

	isSlashBefore := SlashBeforeByUserId(user.ID)

	isIPRegistered := IPExisted(user.RegistrationIp)
	if !isIPRegistered {
		condition2 = true
	}

	isValid = condition1 || (condition2 && !isSlashBefore)
	return
}

func SlashBeforeByUserId(userId int64) (isExisted bool) {
	var count int64
	err := DB.Table("teamup_entries").
		Where("user_id = ?", userId).
		Count(&count).Error

	if err != nil || count > 1 {
		isExisted = true
	}

	return
}
