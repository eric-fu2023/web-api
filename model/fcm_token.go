package model

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm/clause"
	"web-api/util"
)

type FcmToken struct {
	ploutos.FcmTokenC
}

func GetFcmTokens(userIds []int64) ([]FcmToken, error) {
	var fcmTokens []FcmToken
	err := DB.Where("user_id IN ?", userIds).Find(&fcmTokens).Error
	return fcmTokens, err
}

func UpsertFcmToken(c *gin.Context, userId int64, Uuid, fcmToken string) error {
	if fcmToken == "" {
		err := DB.Where("user_id", userId).Where("uuid", Uuid).Delete(&FcmToken{}).Error
		if err != nil {
			util.GetLoggerEntry(c).Errorf("Delete FCM token error: %s", err.Error())
		}
		return err
	}

	ft := FcmToken{
		FcmTokenC: ploutos.FcmTokenC{
			UserId: userId,
			Uuid:   Uuid,
			Token:  fcmToken,
		},
	}
	// Insert or Update on Conflict
	err := DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "uuid"}},
		DoUpdates: clause.AssignmentColumns([]string{"user_id", "token", "deleted_at"}),
	}).Create(&ft).Error
	if err != nil {
		util.GetLoggerEntry(c).Errorf("Insert or update FCM token error: %s", err.Error())
	}
	return err
}
