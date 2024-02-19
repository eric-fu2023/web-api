package model

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	UserAchievementIdFirstAppLoginTutorial     = 1
	UserAchievementIdFirstAppLoginReward       = 2
	UserAchievementIdFirstDepositBonusTutorial = 3
	UserAchievementIdFirstDepositBonusReward   = 4
)

var (
	ErrAchievementAlreadyCompleted = errors.New("user has already completed the achievement")
)

type UserAchievement struct {
	ploutos.UserAchievement
}

type GetUserAchievementCond struct {
	AchievementIds []int64
}

func GetUserAchievements(userId int64, cond GetUserAchievementCond) ([]UserAchievement, error) {
	return GetUserAchievementsWithDB(DB, userId, cond)
}

func GetUserAchievementsWithDB(tx *gorm.DB, userId int64, cond GetUserAchievementCond) ([]UserAchievement, error) {
	if tx == nil {
		return nil, errors.New("tx is nil")
	}

	db := DB.Table(UserAchievement{}.TableName())
	if len(cond.AchievementIds) > 0 {
		db = db.Where("achievement_id IN ?", cond.AchievementIds)
	}

	var achievements []UserAchievement
	err := db.Where("user_id = ?", userId).Find(&achievements).Error
	return achievements, err
}

func CreateUserAchievement(userId int64, achievementId int64) error {
	return CreateUserAchievementWithDB(DB, userId, achievementId)
}

func CreateUserAchievementWithDB(tx *gorm.DB, userId int64, achievementId int64) error {
	if tx == nil {
		return errors.New("tx is nil")
	}

	// Get Achievement
	cond := GetUserAchievementCond{AchievementIds: []int64{achievementId}}
	achievements, err := GetUserAchievementsWithDB(tx.Clauses(clause.Locking{Strength: "UPDATE"}), userId, cond)
	if err != nil {
		return fmt.Errorf("get user achievements err: %w", err)
	}
	if len(achievements) > 0 {
		return ErrAchievementAlreadyCompleted
	}

	ua := UserAchievement{ploutos.UserAchievement{
		UserId:        userId,
		AchievementId: achievementId,
	}}
	return tx.Create(&ua).Error
}
