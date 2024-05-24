package model

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrAchievementAlreadyCompleted = errors.New("user has already completed the achievement")
)

type UserAchievement struct {
	ploutos.UserAchievement
	Achievement *ploutos.Achievement `gorm:"foreignKey:AchievementId;references:ID"`
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
	err := db.Preload("Achievement").
		Where("user_id = ?", userId).
		Find(&achievements).Error
	return achievements, err
}

func CreateUserAchievement(userId int64, achievementId int64) error {
	tx := DB.Begin()
	err := CreateUserAchievementWithDB(tx, userId, achievementId)
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
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

	if len(achievements) > 0 && achievements[0].Achievement == nil {
		return fmt.Errorf("fail to match achievement with id: %d", achievementId)
	}

	if len(achievements) > 0 && !achievements[0].Achievement.AllowRepeat {
		return ErrAchievementAlreadyCompleted
	}

	ua := UserAchievement{UserAchievement: ploutos.UserAchievement{
		UserId:        userId,
		AchievementId: achievementId,
	}}
	return tx.Create(&ua).Error
}

func GetUserAchievementsForMe(userId int64) ([]UserAchievement, error) {
	uaCond := GetUserAchievementCond{AchievementIds: []int64{
		ploutos.UserAchievementIdFirstAppLoginTutorial,
		ploutos.UserAchievementIdFirstAppLoginReward,
		ploutos.UserAchievementIdUpdateBirthday,
	}}
	return GetUserAchievements(userId, uaCond)
}
