package model

import (
	"gorm.io/gorm"
)

func GameProviderUserByProviderAndExternalUser(provider int64, userId string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(`game_provider_id`, provider).Where(`external_user_id`, userId)
	}
}

func ByBrandAgentDeviceAndKey(brand int64, agent int64, device int64, key string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		q := db.Where(`brand_id = ? OR brand_id = 0`, brand).Where(`agent_id = ? OR agent_id = 0`, agent).Where(`platform = ? OR platform = 0`, device)
		if key != "" {
			q = q.Where(`key`, key)
		}
		return q
	}
}

func CategoryTypeWithCategories(db *gorm.DB) *gorm.DB {
	return db.Preload(`Categories`, func(db *gorm.DB) *gorm.DB {
		return db.Order(`sort DESC`)
	})
}

func UserFollowingsByUserId(userId int64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(`user_id`, userId)
	}
}

func UserFollowingsByUserIdAndStreamerId(userId, streamerId int64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Scopes(UserFollowingsByUserId(userId)).Where(`streamer_id`, streamerId)
	}
}
