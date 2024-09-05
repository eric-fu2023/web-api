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
	maxPercentage := int64(10000)
	currentTotalProgress := util.Sum(teamupEntries, func(entry ploutos.TeamupEntry) int64 {
		return entry.FakePercentageProgress
	})

	// IF THIS USER DOESNT FULFILL
	// 1) NEW USER
	// 2) TOP-UP BEFORE
	// THEN SLASH 0%

	isFulfillSlashReq := validSlash(c, user)

	// Check if teamup is shortlisted
	// If yes, then success the current shortlisted
	if isFulfillSlashReq && teamup.ShortlistStatus == ploutos.ShortlistStatusShortlisted {
		_, _ = SuccessShortlisted(teamup, currentTotalProgress, userId)

		// No matter got error or not, need to return
		// No error = success = return
		// Got error = should not continue = return
		return
	}

	if !isFulfillSlashReq {
		beforeProgress = currentTotalProgress
		afterProgress = currentTotalProgress
	} else {
		// 如果currentTotalProgress = 0，beforeProgress = 0，代表第一刀，afterProgress - beforeProgress的差值会比较大
		beforeProgress, afterProgress = GenerateFakeProgress(currentTotalProgress)
	}

	slashEntry := ploutos.TeamupEntry{
		TeamupId: teamupId,
		UserId:   userId,
	}

	slashEntry.TeamupEndTime = teamup.TeamupEndTime
	slashEntry.TeamupCompletedTime = teamup.TeamupCompletedTime
	slashEntry.FakePercentageProgress = afterProgress - beforeProgress

	teamup.TotalFakeProgress = afterProgress

	if isFulfillSlashReq {
		if currentTotalProgress == 0 {
			// Formula
			// (beforeProgress / 100) * (TeamUpTarget / 100) / 100 = ???
			// (6516 / 100) * (21000 / 100) / 100 = $136.836

			slashValue := ((float64(beforeProgress) / 100) * (float64(teamup.TotalTeamUpTarget) / 100)) / 100
			roundedCeilSlashValue := (math.Ceil(slashValue*100) / 100) * 100

			slashEntry.ContributedTeamupDeposit = int64(roundedCeilSlashValue)
			teamup.TotalTeamupDeposit += int64(roundedCeilSlashValue)
		} else {
			teamupContributeFixedAmountString, _ := GetAppConfigWithCache("teamup", "teamup_fixed_amount")
			if teamupContributeFixedAmountString != "" {
				contributeAmount, _ := strconv.Atoi(teamupContributeFixedAmountString)
				slashEntry.ContributedTeamupDeposit = int64(contributeAmount)
				teamup.TotalTeamupDeposit += int64(contributeAmount)
			}
		}
	}

	// IF THIS TOTAL_TEAMUP_DEPOSIT >= TOTAL_TEAMUP_TARGET
	// MEANS SUCCESS, WILL THEN CHECK IF CURRENT TERM ALREADY HAS 20
	// IF ALREADY HAS 20, PICK 4 LOWEST AMOUNT AND FLAG AS QUALIFIED
	// IGNORE NOT QUALIFIED, ONLY THE FIRST THAT INVITED 1 MORE WILL SUCCESS
	if teamup.TotalTeamupDeposit >= teamup.TotalTeamUpTarget {

		currentTerm, _ := GetCurrentTermNum()
		teamupTermSizeString, _ := GetAppConfigWithCache("teamup", "max_slash_amount")
		termSize, _ := strconv.Atoi(teamupTermSizeString)

		afterProgress = maxPercentage
		slashEntry.FakePercentageProgress = afterProgress - beforeProgress
		teamup.TotalFakeProgress = afterProgress

		// Check If Term 2 Already Has termSize - 1 (20 - 1 = 19)
		termTeamups, _ := FindExceedTargetStatusPendingByTerm(currentTerm)
		teamup.Term = currentTerm
		if len(termTeamups) > termSize-1 {
			teamup.Term++
		}
		if len(termTeamups) == termSize-1 {
			// If term has termSize-1, means can put assign term to one more teamup
			// Calculate immediately since 19 + current = 20
			termTeamups = append(termTeamups, teamup)
			sort.Slice(termTeamups, func(i, j int) bool {
				return int(termTeamups[i].TotalTeamUpTarget) < int(termTeamups[j].TotalTeamUpTarget)
			})
			termTeamups = termTeamups[:4]
			var ids []int64
			for _, t := range termTeamups {
				ids = append(ids, t.ID)
			}

			// Turn 4 lowest slash target amount into shortlisted
			err = FlagStatusShortlisted(ids)
			if err != nil {
				return
			}
		}
	}

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

	isIPRegistered := IPExisted(user.RegistrationIp)
	if !isIPRegistered {
		condition2 = true
	}

	isValid = condition1 || condition2
	return
}
