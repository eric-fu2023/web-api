package model

import (
	"context"
	"fmt"
	"log"
	"math"
	"sort"
	"strconv"
	"time"
	"web-api/conf/consts"

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
	ContributedAmount    float64   `json:"contributed_amount"`
	ContributedTime      time.Time `json:"contributed_time"`
	Nickname             string    `json:"nickname"`
	Avatar               string    `json:"avatar"`
	Progress             int64     `json:"progress"`
	AdjustedFiatProgress float64   `json:"adjusted_fiat_progress"`
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

func GetAllTeamUpEntries(brand int, teamupId int64, page, limit int) (res TeamupEntryCustomRes, err error) {

	teamup, err := GetTeamUpByTeamUpId(teamupId)
	if err != nil {
		return
	}

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

	res, _ = FormatAdjustedFiatProgress(brand, res, teamup)

	return
}

func CreateSlashBetRecord(c *gin.Context, teamupId int64, user ploutos.User, i18n i18n.I18n) (teamup ploutos.Teamup, isSuccess bool, wonTeamupIds []int64, err error) {

	brand := c.MustGet(`_brand`).(int)
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
	// if teamup.Term != 0 && teamup.ShortlistStatus != ploutos.ShortlistStatusNotShortlist {
	// 	// 如果有Term，如果这单是成功 / 入选
	// 	if isValidSlash {
	// 		isSuccessShortlisted, _ := SuccessShortlisted(brand, teamup, currentTotalProgress, userId)

	// 		// No matter got error or not, need to return
	// 		// No error = success = return
	// 		// Got error = should not continue = return
	// 		if isSuccessShortlisted {
	// 			isTeamupSuccess = true
	// 			isSuccess = true
	// 			return
	// 		}

	// 		// Give random percentage if shortlisted by others, slash for fun
	// 		shouldGiveRandomPercentage = true
	// 	}
	// } else if teamup.Term != 0 && teamup.ShortlistStatus == ploutos.ShortlistStatusNotShortlist {

	// 	// 如果有Term，就代表CONTRIBUTION >= TARGET+不是入选/成功，就意思意思砍0
	// 	isValidSlash = false
	// 	shouldGiveRandomPercentage = true
	// }

	if teamup.Term != 0 && teamup.ShortlistStatus == ploutos.ShortlistStatusNotShortlist {
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

	err = DB.Transaction(func(tx *gorm.DB) (err error) {
		// 如果未进入候选池 + 后端数值已达标 = 加入候选池 （20进4）
		if isValidSlash && teamup.Term == 0 && teamup.TotalTeamupDeposit >= teamup.TotalTeamUpTarget {

			// afterProgress = int64(9999)
			slashEntry.FakePercentageProgress = afterProgress - beforeProgress
			teamup.TotalFakeProgress = afterProgress

			// 获得该期
			currentTerm, _ := GetCurrentTermNum(brand, teamup.BetReportGameType)
			teamupTermSizeString, _ := GetAppConfigWithCache("teamup", "term_size")
			termSize, _ := strconv.Atoi(teamupTermSizeString)
			termTeamups, _ := FindExceedTargetByTerm(brand, currentTerm, teamup.BetReportGameType)
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
						teamup.Status = int(ploutos.TeamupStatusSuccess)
						teamup.ShortlistStatus = ploutos.ShortlistStatusShortlistWin
						teamup.TotalFakeProgress = int64(10000)
						teamup.TeamupCompletedTime = time.Now().UTC().Unix()
					}
					if t.Status == int(ploutos.TeamupStatusFail) {
						continue
					}
					ids = append(ids, t.ID)
				}

				// 选价值最小的4张单晋级，之后这4张单选一张砍单成功
				err = FlagStatusShortlistedWin(tx, ids)
				if err != nil {
					return
				}

				wonTeamupIds = ids
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
			return
		}

		return
	})

	isSuccess = true

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

func TeamupEntriesByDateRange(c context.Context, userId int64, startDate, endDate time.Time) (teamupEntries []ploutos.TeamupEntry, err error) {

	db := DB.Table("teamup_entries").Where("user_id", userId).Where("contributed_teamup_deposit != ?", 0)
	db.Where("created_at >= ?", startDate)
	db.Where("created_at < ?", endDate)

	err = db.Find(&teamupEntries).Error

	return
}

func validSlash(c *gin.Context, user ploutos.User) (isValid bool) {

	var condition bool
	// topUpCashOrder, _ := FirstTopup(c, user.ID)
	// if topUpCashOrder.ID != "" {
	// 	condition1 = true
	// }

	isSlashBefore := SlashBeforeByUserId(user.ID)

	isIPRegistered := IPExisted(user.RegistrationIp)
	if !isIPRegistered {
		condition = true
	}

	if condition && !isSlashBefore {
		isValid = true
		return
	}

	// isValid = condition1 || (condition2 && !isSlashBefore)

	// if isValid == false {
	// 	return
	// }

	const hours = 48
	start := time.Now().Add(-hours * time.Hour).UTC()

	// const minutes = 5
	// start := time.Now().Add(-(minutes) * time.Hour).UTC()

	end := time.Now().UTC()

	topupRecords, _ := TopupsByDateRange(c, user.ID, start, end)

	if len(topupRecords) == 0 {
		return
	}

	teamupEntries, _ := TeamupEntriesByDateRange(c, user.ID, start, end)

	if len(topupRecords) > len(teamupEntries) {
		isValid = true
	}

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

func ShouldPopRoulette(brandId int, userId int64) (shouldPop bool) {

	_, _, spinGameTypes := GetTeamUpGameTypeSliceByBrand(brandId)

	// condition1 := true
	condition2 := true

	// 1）新用户（没有注册过的ip和uuid）+ 超过某个时间点注册
	// 2）砍单记录里没启动过轮盘砍单（发起过轮盘砍单就不可以再发起了）

	// 新用户（没有注册过的ip和uuid）+ 超过某个时间点注册
	// var countUsers int64
	// _ = DB.Table("users").
	// 	Where("id = ?", userId).
	// 	Where("created_at > ?", "2024-09-20 00:00:00.000000").
	// 	Count(&countUsers).Error

	// if countUsers > 0 {
	// 	isIPRegistered := IPExisted(ip)
	// 	// IP 已注册过
	// 	if isIPRegistered {
	// 		condition1 = false
	// 	}
	// } else {
	// 	// 老用户
	// 	condition1 = false
	// }

	// 砍单记录里没启动过轮盘砍单（发起过轮盘砍单就不可以再发起了）
	var countStartedSpin int64
	_ = DB.Table("teamups").
		Where("user_id = ?", userId).
		Where("bet_report_game_type in ?", spinGameTypes).
		Count(&countStartedSpin).Error

	if countStartedSpin > 0 {
		// 已发起过轮盘砍单
		condition2 = false
	}

	// return condition1 && condition2
	return condition2
}

func FormatAdjustedFiatProgress(brand int, teamupEntries TeamupEntryCustomRes, teamup ploutos.Teamup) (res TeamupEntryCustomRes, totalFiatProgress float64) {
	switch brand {
	case consts.BrandAha:
		partialTotalProgress := 0.00

		// teamupEntries = mapFormatAdjustedFiatProgress(teamupEntries, func(entry any) any {
		// 	floorFiatProgress := math.Floor((float64(teamupEntries[i].Progress)/10000)*float64(teamup.TotalTeamUpTarget)/100*100) / 100
		// 	entry.AdjustedFiatProgress = floorFiatProgress
		// })

		teamupEntries = mapFormatAdjustedFiatProgress(teamupEntries, func(entries TeamupEntryCustomRes) TeamupEntryCustomRes {
			for i := len(teamupEntries) - 1; i >= 0; i-- {

				floorFiatProgress := math.Floor((float64(teamupEntries[i].Progress)/10000)*float64(teamup.TotalTeamUpTarget)/100*100) / 100
				teamupEntries[i].AdjustedFiatProgress = floorFiatProgress

				if i != 0 {
					partialTotalProgress += teamupEntries[i].AdjustedFiatProgress
				}
			}

			return teamupEntries

		})

		if teamup.TotalFakeProgress >= 10000 {
			teamupEntries[0].AdjustedFiatProgress = math.Floor((float64(teamup.TotalTeamUpTarget)/100-float64(partialTotalProgress))*100) / 100
		}

	case consts.BrandBatAce:
		partialTotalProgress := 0.00

		// teamupEntries = mapFormatAdjustedFiatProgress(teamupEntries, func(entry any) any {
		// 	floorFiatProgress := math.Floor((float64(teamupEntries[i].Progress)/10000)*float64(teamup.TotalTeamUpTarget)/100*100) / 100
		// 	entry.AdjustedFiatProgress = floorFiatProgress
		// })

		teamupEntries = mapFormatAdjustedFiatProgress(teamupEntries, func(entries TeamupEntryCustomRes) TeamupEntryCustomRes {
			for i := len(teamupEntries) - 1; i >= 0; i-- {

				floorFiatProgress := math.Floor((float64(teamupEntries[i].Progress)/10000)*float64(teamup.TotalTeamUpTarget)/100*100) / 100
				teamupEntries[i].AdjustedFiatProgress = floorFiatProgress

				if int(floorFiatProgress) == 0 {
					teamupEntries[i].AdjustedFiatProgress = 1
				}

				teamupEntries[i].AdjustedFiatProgress = float64(int(teamupEntries[i].AdjustedFiatProgress))

				if len(teamupEntries) == 1 {
					break
				}

				if i != 0 {
					partialTotalProgress += teamupEntries[i].AdjustedFiatProgress
				} else {

					if (teamup.Status == int(ploutos.TeamupStatusPending) || teamup.Status == int(ploutos.TeamupStatusFail)) && partialTotalProgress >= float64(int(teamup.TotalTeamUpTarget/100)-1) {
						teamupEntries[i].AdjustedFiatProgress = 0
					}
				}

				if int(partialTotalProgress) >= int(teamup.TotalTeamUpTarget/100) {
					prev := partialTotalProgress - teamupEntries[i].AdjustedFiatProgress
					teamupEntries[i].AdjustedFiatProgress = float64(teamup.TotalTeamUpTarget/100-1) - float64(prev)
					partialTotalProgress = float64(int(teamup.TotalTeamUpTarget/100) - 1)
				}

			}

			if teamup.TotalFakeProgress >= 10000 {
				teamupEntries[0].AdjustedFiatProgress = math.Floor((float64(teamup.TotalTeamUpTarget)/100-float64(partialTotalProgress))*100) / 100
			}

			return teamupEntries

		})
	}

	res = teamupEntries

	return
}

func UpdateLastTeamupEntryToMaxProgress(teamupIds []int64) {
	for _, teamupId := range teamupIds {
		updateErr := DB.Transaction(func(tx *gorm.DB) (err error) {

			teamupEntries, err := FindTeamupEntryByTeamupId(teamupId)
			if err != nil {
				return
			}
			currentTotalProgress := util.Sum(teamupEntries, func(entry ploutos.TeamupEntry) int64 {
				return entry.FakePercentageProgress
			})

			remainingProgress := int64(10000) - currentTotalProgress

			teamupEntries[len(teamupEntries)-1].FakePercentageProgress += remainingProgress

			err = tx.Save(&teamupEntries[len(teamupEntries)-1]).Error

			return
		})

		if updateErr != nil {
			util.Log().Error(`Update Last Teamup Entry To Max Progress Error - %v`, updateErr)
			return
		}
	}
}

func mapFormatAdjustedFiatProgress[T any](t T, f func(T) T) T {
	return f(t)
}
