package model

import (
	"gorm.io/gorm"
	"time"
)

func GameVendorUserByVendorAndExternalUser(vendor int64, userId string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(`game_vendor_id`, vendor).Where(`external_user_id`, userId)
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

func ByOrderListConditions(userId int64, isParlay bool, isSettled bool, start time.Time, end time.Time) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		db.Where(`user_id`, userId).Where(`is_parlay`, isParlay)
		if isSettled {
			db.Where(`reward_status`, 5)
		} else {
			db.Where(`reward_status != ?`, 5)
		}
		if !start.IsZero() && !end.IsZero() {
			db.Where(`bet_time >= ?`, start).Where(`bet_time <= ?`, end)
		}
		return db
	}
}
