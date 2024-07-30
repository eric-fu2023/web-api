package service

import (
	"context"
	"fmt"
	"strconv"
	"time"
	"web-api/cache"
	"web-api/model"
	"web-api/serializer"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

type PopupService struct {
}

func (service *PopupService) buildKey(userId int64) (key string) {
	now := time.Now()
	return "popup_records/"+strconv.FormatInt(userId,10) +"/"+ now.Format("2006-01-02")
}

func (service *PopupService) ShowPopup(c *gin.Context) (r serializer.Response, err error) {
	type PopupResponse struct {
		Type int         `json:"type"`
		Data interface{} `json:"data"`
	}

	u, _ := c.Get("user")
	user := u.(model.User)
	// check what to popup
	// check redis which one has been popup
	key := service.buildKey(user.ID)
	res := cache.RedisClient.Get(context.Background(), key)
	if res.Err() != nil && res.Err() != redis.Nil {
		// if redis get error, return error
		fmt.Print("Redis Get failed, ",res.Err())
		return r, res.Err()
	} else if res.Err() == redis.Nil {
		// if no display record found in redis, start finding the popup window
		shouldPopupWinLose, err := model.ShouldPopupWinLose(user)
		if err != nil {
			return r, err
		}
		if shouldPopupWinLose {
			var service WinLoseService
			data, err := service.Get(c)
			r.Data = PopupResponse{
				Type: 1,
				Data: data,
			}
			return r, err
		}

		ShouldVIP, err := model.ShouldPopupVIP(user)
		if err != nil {
			return r, err
		}
		if ShouldVIP {
			var service VipService
			data, err := service.Get(c)
			r.Data = PopupResponse{
				Type: 3,
				Data: data,
			}
			return r, err
		}
	} else {
		redisPopup, err := strconv.Atoi(res.Val())
		if err != nil && err != redis.Nil {
			return r, err
		}
		if redisPopup < 2 {
			shouldPopupWinLose, err := model.ShouldPopupWinLose(user)
			if err != nil {
				return r, err
			}
			if shouldPopupWinLose {
				var service WinLoseService
				data, err := service.Get(c)
				r.Data = PopupResponse{
					Type: 1,
					Data: data,
				}
				return r, err
			}
		}
		if redisPopup < 3 {
			// TODO: this is for Kan Dan
			// shouldPopupWinLose, err := model.ShouldPopupWinLose(user)
			// if shouldPopupWinLose {
			// 	var service WinLoseService
			// 	return service.Get(c)
			// }
		}
		if redisPopup < 4 {
			ShouldVIP, err := model.ShouldPopupVIP(user)
			if err != nil {
				return r, err
			}
			if ShouldVIP {
				var service VipService
				data, err := service.Get(c)
				r.Data = PopupResponse{
					Type: 3,
					Data: data,
				}
				return r, err
			}
		}
	}

	return
}
