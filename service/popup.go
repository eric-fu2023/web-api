package service

import (
	"context"
	"fmt"
	"strconv"
	"time"
	"web-api/cache"
	"web-api/model"
	"web-api/serializer"

	models "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

type PopupService struct {
	Condition int64 `form:"condition" json:"condition"`
}

func (service *PopupService) buildKey(userId int64) (key string) {
	now := time.Now()
	return "popup_records/"+strconv.FormatInt(userId,10) +"/"+ now.Format("2006-01-02")
}

func (service *PopupService) ShowPopup(c *gin.Context) (r serializer.Response, err error) {
	type PopupResponse struct {
		Type int         `json:"type"`
		CanFloat bool `json:"can_float"`
		Data interface{} `json:"data"`
	}

	u, _ := c.Get("user")
	user := u.(model.User)
	// check what to popup
	PopupTypes, err := model.GetPopupList(service.Condition)

	// check redis which one has been popup
	key := "popup/records/" + time.Now().Format("2006-01-02")
	res := cache.RedisClient.HGet(context.Background(), key,strconv.FormatInt(user.ID, 10))
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
				CanFloat: WinLoseFloat(PopupTypes),
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
			if err != nil {
				return r, err
			}
			r.Data = PopupResponse{
				Type: 3,
				CanFloat: VIPFloat(PopupTypes),
				Data: data,
			}
			return r, err
		}
	} else {
		redisPopup, err := strconv.Atoi(res.Val())
		if err != nil && err != redis.Nil {
			return r, err
		}
		// this case should nvr happen because the less value is 1 in redis
		if redisPopup < 1 {
			shouldPopupWinLose, err := model.ShouldPopupWinLose(user)
			if err != nil {
				return r, err
			}
			if shouldPopupWinLose  && WinLoseAvailable(PopupTypes){
				var service WinLoseService
				data, err := service.Get(c)
				r.Data = PopupResponse{
					Type: 1,
					CanFloat: WinLoseFloat(PopupTypes),
					Data: data,
				}
				return r, err
			}
		}
		//----------------------------------------------------------------------
		if redisPopup < 2 {
			// TODO: this is for Kan Dan
			// shouldPopupWinLose, err := model.ShouldPopupWinLose(user)
			// if shouldPopupWinLose {
			// 	var service WinLoseService
			// 	return service.Get(c)
			// }
		}
		if redisPopup < 3 {
			ShouldVIP, err := model.ShouldPopupVIP(user)
			if err != nil {
				return r, err
			}
			if ShouldVIP && VIPAvailable(PopupTypes){
				var service VipService
				data, err := service.Get(c)
				r.Data = PopupResponse{
					Type: 3,
					CanFloat: VIPFloat(PopupTypes),
					Data: data,
				}
				return r, err
			}
		}
		if redisPopup < 4 {
			// TODO: this is for spins
			// shouldPopupWinLose, err := model.ShouldPopupWinLose(user)
			// if shouldPopupWinLose {
			// 	var service WinLoseService
			// 	return service.Get(c)
			// }
		}
	}
	r.Msg = "no popup available"
	r.Data = PopupResponse{Type:-1}
	return
}


func WinLoseAvailable(popups []models.Popups) bool {
    for _, popup := range popups {
        if popup.PopupType == 1 {
            return true // Found a popup with PopupType == 1
        }
    }
    return false // No popup with PopupType == 1 was found
}

func KanDanAvailable(popups []models.Popups) bool {
    for _, popup := range popups {
        if popup.PopupType == 2 {
            return true // Found a popup with PopupType == 2
        }
    }
    return false // No popup with PopupType == 2 was found
}
func VIPAvailable(popups []models.Popups) bool {
    for _, popup := range popups {
        if popup.PopupType == 3 {
            return true // Found a popup with PopupType == 3
        }
    }
    return false // No popup with PopupType == 3 was found
}
func SpinAvailable(popups []models.Popups) bool {
    for _, popup := range popups {
        if popup.PopupType == 4 {
            return true // Found a popup with PopupType == 4
        }
    }
    return false // No popup with PopupType == 4 was found
}


func WinLoseFloat(popups []models.Popups) bool {
    for _, popup := range popups {
        if popup.PopupType == 1 {
            return popup.CanFloat // Found a popup with PopupType == 1
        }
    }
    return false // No popup with PopupType == 1 was found
}

func KanDanFloat(popups []models.Popups) bool {
    for _, popup := range popups {
        if popup.PopupType == 2 {
            return popup.CanFloat // Found a popup with PopupType == 2
        }
    }
    return false // No popup with PopupType == 2 was found
}
func VIPFloat(popups []models.Popups) bool {
    for _, popup := range popups {
        if popup.PopupType == 3 {
            return popup.CanFloat // Found a popup with PopupType == 3
        }
    }
    return false // No popup with PopupType == 3 was found
}
func SpinFloat(popups []models.Popups) bool {
    for _, popup := range popups {
        if popup.PopupType == 4 {
            return popup.CanFloat // Found a popup with PopupType == 4
        }
    }
    return false // No popup with PopupType == 4 was found
}