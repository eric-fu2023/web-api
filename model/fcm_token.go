package model

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"web-api/util"
)

type FcmToken struct {
	ploutos.FcmToken
}

func GetFcmTokenStrings(userIds []int64) ([]string, error) {
	var fcmTokens []FcmToken
	fcmTokens, err := getFcmTokens(userIds)
	if err != nil {
		fmt.Println("Get fcm tokens error:", err.Error())
		return nil, err
	}
	fcmTokenStr := make([]string, len(fcmTokens))
	for i, fcmToken := range fcmTokens {
		fcmTokenStr[i] = fcmToken.Token
	}
	return fcmTokenStr, nil
}

func getFcmTokens(userIds []int64) ([]FcmToken, error) {
	var fcmTokens []FcmToken
	err := DB.Where("user_id IN ?", userIds).Find(&fcmTokens).Error
	return fcmTokens, err
}

func UpsertFcmToken(c *gin.Context, userId int64, uuid, fcmToken string) error {
	if fcmToken == "" {
		err := DB.Where("user_id", userId).Where("uuid", uuid).Delete(&FcmToken{}).Error
		if err != nil {
			util.GetLoggerEntry(c).Errorf("Delete FCM token error: %s", err.Error())
		}
		return err
	}

	err := DB.Transaction(func(tx *gorm.DB) error {
		// Get existing
		var existing FcmToken
		err := tx.Unscoped().
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Table(FcmToken{}.TableName()).
			Where("uuid = ?", uuid).
			First(&existing).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			util.GetLoggerEntry(c).Errorf("Get existing FCM token error: %s", err.Error())
			return err
		}

		// Update if there is existing
		if err == nil {
			err = tx.Unscoped().Model(&FcmToken{}).Where("id", existing.ID).
				Updates(map[string]any{
					"user_id":    userId,
					"token":      fcmToken,
					"deleted_at": nil,
				}).Error
			if err != nil {
				util.GetLoggerEntry(c).Errorf("Update FCM token error: %s", err.Error())
				return err
			}
			return nil
		}

		// Create
		ft := FcmToken{
			FcmToken: ploutos.FcmToken{
				UserId: userId,
				Uuid:   uuid,
				Token:  fcmToken,
			},
		}

		err = tx.Create(&ft).Error
		if err != nil {
			util.GetLoggerEntry(c).Errorf("Create FCM token error: %s", err.Error())
			return err
		}

		return nil
	})
	if err != nil {
		util.GetLoggerEntry(c).Errorf("Transaction error: %s", err.Error())
		return err
	}

	return err
}
