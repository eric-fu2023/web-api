package model

import (
	"gorm.io/gorm"
)

func GameProviderUserByProviderAndExternalUser(provider int64, userId string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(`game_provider_id`, provider).Where(`external_user_id`, userId)
	}
}

func ByBrandAgentPlatformAndKey(brand int64, agent int64, platform int64, key string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		q := db.Scopes(ByBrandAgentAndPlatform(brand, agent, platform))
		if key != "" {
			q = q.Where(`key`, key)
		}
		return q
	}
}

func CategoryTypeWithCategories(db *gorm.DB) *gorm.DB {
	return db.Preload(`Categories`, func(db *gorm.DB) *gorm.DB {
		return db.Scopes(Sort)
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

func ByBrandAgentAndPlatform(brand int64, agent int64, platform int64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(`brand_id = ? OR brand_id = 0`, brand).Where(`agent_id = ? OR agent_id = 0`, agent).Where(`platform = ? OR platform = 0`, platform)
	}
}

func Sort(db *gorm.DB) *gorm.DB {
	return db.Order(`sort DESC`)
}

func SortByCreated(db *gorm.DB) *gorm.DB {
	return db.Order(`created_at DESC`)
}

func ByUserId(userId int64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(`user_id`, userId).Limit(1)
	}
}

func ByIds(ids []int64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(`id`, ids)
	}
}
