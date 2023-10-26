package model

import (
	"fmt"
	"gorm.io/gorm"
	"strings"
	"time"
	"web-api/conf/consts"
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

func SortById(db *gorm.DB) *gorm.DB {
	return db.Order(`id`)
}

func ByStatus(db *gorm.DB) *gorm.DB {
	return db.Where(`status`, 1)
}

func ByUserId(userId int64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(`user_id`, userId).Limit(1)
	}
}

func ByUserStatusAndRole(status []int64, role []int64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(`status`, status).Where(`role`, role)
	}
}

func ByActiveNonStreamerUser(db *gorm.DB) *gorm.DB {
	return ByUserStatusAndRole([]int64{1}, []int64{consts.UserRole["user"], consts.UserRole["test_user"]})(db)
}

func ByIds(ids []int64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(`id`, ids)
	}
}

func ByOrderListConditions(userId int64, gameType []int64, isParlay bool, isSettled bool, start time.Time, end time.Time) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		db.Where(`user_id`, userId).Where(`is_parlay`, isParlay).Where(`game_type`, gameType)
		if isSettled {
			db.Where(`status`, 5)
		} else {
			db.Where(`status != ?`, 5)
		}
		if !start.IsZero() && !end.IsZero() {
			db.Where(`bet_time >= ?`, start).Where(`bet_time <= ?`, end)
		}
		return db
	}
}

func ByBetTimeSort(db *gorm.DB) *gorm.DB {
	return db.Order(`bet_time DESC`)
}

func ByPlatformExpended(platform int64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if platform == consts.Platform["pc"] {
			db.Where(`web`, 1)
		} else if platform == consts.Platform["h5"] {
			db.Where(`h5`, 1)
		} else if platform == consts.Platform["android"] {
			db.Where(`android`, 1)
		} else if platform == consts.Platform["ios"] {
			db.Where(`ios`, 1)
		}
		return db
	}
}

func ByCategoryAndBrand(categoryId int64, brandId int64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		db.Where(`brand_id = ? OR brand_id = 0`, brandId)
		if categoryId != 0 {
			db.Where(`category_id`, categoryId)
		}
		return db
	}
}

func ByDcRoundAndWager(roundId string, wagerId string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		db.Where(`round_id`, roundId).Where(`wager_id`, wagerId)
		return db
	}
}

func ByDcRoundWagerAndWagerType(roundId string, wagerId string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		db.Scopes(ByDcRoundAndWager(roundId, wagerId)).Where(`wager_type != 0`)
		return db
	}
}

func ByGameIdsBrandAndIsFeatured(gameIds []string, brandId int64, isFeatured bool) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		db.Where(`sub_game_brand.brand_id = ? OR sub_game_brand.brand_id = 0`, brandId)
		if len(gameIds) > 0 {
			db.Where(`sub_game_brand.id`, gameIds).Order(fmt.Sprintf(`ARRAY_POSITION(ARRAY[%s], sub_game_brand.id)`, strings.Join(gameIds, ",")))
		} else {
			db.Order(`sub_game_brand.sort`)
		}
		if isFeatured {
			db.Where(`sub_game_brand.is_featured`, true)
		}
		return db
	}
}

func ByPlatformAndStatusOfSubAndVendor(platform int64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		db.Joins(`JOIN game_vendor_brand gvb ON sub_game_brand.vendor_brand_id = gvb.id`).
			Where(`gvb.status`, 1).Where(`sub_game_brand.status`, 1)
		if platform == consts.Platform["pc"] {
			db.Where(`gvb.web`, 1).Where(`sub_game_brand.web`, 1)
		} else if platform == consts.Platform["h5"] {
			db.Where(`gvb.web`, 1).Where(`sub_game_brand.h5`, 1)
		} else if platform == consts.Platform["android"] {
			db.Where(`gvb.web`, 1).Where(`sub_game_brand.android`, 1)
		} else if platform == consts.Platform["ios"] {
			db.Where(`gvb.web`, 1).Where(`sub_game_brand.ios`, 1)
		}
		return db
	}
}

func ByMaintenance(db *gorm.DB) *gorm.DB {
	now := time.Now()
	db.Where(`gvb.start_time IS NULL OR gvb.start_time = '0001-01-01' OR ? NOT BETWEEN gvb.start_time AND gvb.end_time`, now).
		Where(`sub_game_brand.start_time IS NULL OR sub_game_brand.start_time = '0001-01-01' OR ? NOT BETWEEN sub_game_brand.start_time AND sub_game_brand.end_time`, now)
	return db
}

func UserFavouriteByUserIdTypeAndSportId(userId, t, sportId int64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		db.Where(`user_id`, userId).Where(`type`, t)
		if sportId != 0 {
			db.Where(`sport_id`, sportId)
		}
		return db
	}
}

func UserFavouriteByUserIdTypeGameIdAndSportId(userId, t, gameId, sportId int64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Scopes(UserFavouriteByUserIdTypeAndSportId(userId, t, sportId)).Where(`game_id`, gameId)
	}
}
