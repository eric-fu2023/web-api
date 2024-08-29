package model

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
	"web-api/cache"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm/logger"
)

func ShouldPopupWinLose(user User) (bool, error) {
	now := time.Now()
	key := "popup/win_lose/"+now.Format("2006-01-02")
	// here we need to use db2 to get the task system redis data
	res := cache.RedisClient.HGet(context.Background(), key, strconv.FormatInt(user.ID, 10))
	if res.Err() != nil {
		if res.Err() == redis.Nil{
			return false, nil
		}else{
			return false, res.Err()
		}
	}
	return true, nil
}

func ShouldPopupTeamUp(user User) (bool, error) {
	now := time.Now()
	TodayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yesterdayStart := time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, now.Location())
	var team_up ploutos.Teamup
	// status = 2 is success,    status = 0 is onging
	err := DB.Model(ploutos.Teamup{}).Where("user_id = ? AND updated_at < ? AND updated_at > ? AND status in (1,0) AND total_fake_progress !=0", user.ID, TodayStart, yesterdayStart).Order("status DESC, total_teamup_deposit DESC").First(&team_up).Error
	if errors.Is(err, logger.ErrRecordNotFound) {
			err = nil
			// if no team up record, we return nil
			return false, err
		}
		if err != nil {
			fmt.Println("ShouldPopupTeamUp teamup err", err.Error())
			return false, err
		}

	return true, nil
}


func ShouldPopupVIP(user User) (bool, error) {
	key:="popup/vip"
	res:=cache.RedisClient.HGet(context.Background(), key, strconv.FormatInt(user.ID, 10))
	vip, err := GetVipWithDefault(nil, user.ID)
	current_vip_level := vip.VipRule.VIPLevel

	if res.Err() == redis.Nil {
		fmt.Println("vip_level not found in redis")
		// if no vip level up record, we check if user vip level is more than 0
		if current_vip_level > 0{
			res := cache.RedisClient.HSet(context.Background(), key, strconv.FormatInt(user.ID, 10), vip.VipRule.VIPLevel)
			if res.Err() != nil{
				fmt.Println("update downgrade vip level failed, ", res.Err().Error())
			}
			return true, nil
		}else{
			fmt.Println("current_vip_level = 0")
			return false, nil
		}
	}
	if res.Err() != nil && res.Err() != redis.Nil{
		fmt.Println("get vip redis error")
		return false, res.Err()
	}
	previous_vip_level, err := strconv.ParseInt(res.Val(),10,64)
	if err != nil {
		fmt.Println("err convert vip level from redis string to int64, ", err)
	}
	if current_vip_level > previous_vip_level {
			res:=cache.RedisClient.HSet(context.Background(), key, strconv.FormatInt(user.ID, 10), vip.VipRule.VIPLevel)
			if res.Err() != nil{
				fmt.Println("update downgrade vip level failed, ", res.Err().Error())
			}
			return true, nil
		} else if current_vip_level < previous_vip_level{
			// if there is a vip downgrade, we need to update the deleted_at for the record
			res:=cache.RedisClient.HSet(context.Background(), key, strconv.FormatInt(user.ID, 10), vip.VipRule.VIPLevel)
			if res.Err() != nil{
				fmt.Println("update downgrade vip level failed, ", res.Err().Error())
			}
		}
	
	return false,err
}

// here we only check if user has remaining counts.
func ShouldPopupSpin(user User, spin_promotion_id int) (bool, error) {
	// need to check if user has used all spin chances
	now := time.Now()
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	var previous_spin_result []ploutos.SpinResult
	err := DB.Model(ploutos.SpinResult{}).Joins("LEFT JOIN spins ON spins.id = spin_results.spin_id").Where("user_id = ? AND spins.promotion_id = ?", user.ID, spin_promotion_id).Where("created_at > ?", startOfToday).Order("created_at DESC").Find(&previous_spin_result).Error

	if err==nil || errors.Is(err, logger.ErrRecordNotFound) {
		fmt.Println("get spin result error, but the error is no records",err)
		// if spin result not found
		err = nil
		Shown(user)
		return true, nil
	}
	var spin ploutos.Spins
	err = DB.Model(ploutos.Spins{}).Where("promotion_id = ?", spin_promotion_id).Find(&spin).Error
	// if not displayed today
	if len(previous_spin_result) < spin.Counts {
		Shown(user)
		return true, nil
	}
	return false, nil

}

func GetPopupList(condition int64) (resp_list []ploutos.Popups, err error) {
	err = DB.Debug().Model(ploutos.Popups{}).Where("condition = ?", condition).
		Find(&resp_list).Error
	return
}

func Shown(user User) ( err error) {
	key := "popup/records/" + time.Now().Format("2006-01-02")
	res := cache.RedisClient.HSet(context.Background(), key, user.ID, "5")
	if res.Err() != nil {
		fmt.Print("insert win lose popup record into redis failed ", key)
	}
	expire_time, err := strconv.Atoi(os.Getenv("POPUP_RECORD_EXPIRE_MINS"))
	cache.RedisClient.ExpireNX(context.Background(), key, time.Duration(expire_time)*time.Minute)
	return
}