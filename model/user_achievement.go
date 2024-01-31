package model

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"gorm.io/gorm"
)

const (
	UserAchievementIdFirstAppLoginTutorial = 1
	UserAchievementIdFirstAppLoginReward   = 2
)

type UserAchievement struct {
	ploutos.UserAchievement
}

type GetUserAchievementCond struct {
	AchievementIds []int64
}

func GetUserAchievements(userId int64, cond GetUserAchievementCond) ([]UserAchievement, error) {
	db := DB.Table(UserAchievement{}.TableName())
	if len(cond.AchievementIds) > 0 {
		db = db.Where("achievement_id IN ?", cond.AchievementIds)
	}

	var achievements []UserAchievement
	err := db.Where("user_id = ?", userId).Find(&achievements).Error
	return achievements, err
}

func CreateUserAchievement(userId int64, achievementId int64) error {
	ua := UserAchievement{ploutos.UserAchievement{
		UserId:        userId,
		AchievementId: achievementId,
	}}
	return DB.Create(&ua).Error
}

func CreateUserAchievementWithDB(tx *gorm.DB, userId int64, achievementId int64) error {
	ua := UserAchievement{ploutos.UserAchievement{
		UserId:        userId,
		AchievementId: achievementId,
	}}
	return tx.Create(&ua).Error
}
