package service

import (
	"context"
	"fmt"
	"os"
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
type PopupSpinId struct {
	SpinId int `json:"spin_id"`
}
type PopupResponse struct {
	UserId int64        `json:"user_id"`
	Type   int          `json:"type"`
	Float  []PopupFloat `json:"float"`
	Data   interface{}  `json:"data"`
}

type PopupFloat struct {
	Type     int64  `json:"type"`
	Id       int    `json:"id"`
	FloatUrl string `json:"float_url"`
}

func (service *PopupService) ShowPopup(c *gin.Context) (r serializer.Response, err error) {
	// check what to popup
	PopupTypes, err := model.GetPopupList(service.Condition)
	if err != nil {
		fmt.Println("get PopupTypes error", err)
	}

	u, isUser := c.Get("user")
	if !isUser {
		// if not login, show spin only
		should_spin, spin_promotion_id := SpinAvailable(PopupTypes)
		spin_promotion_id_string := strconv.Itoa(spin_promotion_id)
		spin, _ := model.GetSpinByPromotionId(spin_promotion_id_string)
		var floats []PopupFloat

		if should_spin {
			floats = append(floats, PopupFloat{
				Type: 5,
				Id:   spin_promotion_id,
				FloatUrl: serializer.Url(spin.FloatUrl),
			})
			spin_id_data := PopupSpinId{
				SpinId: spin_promotion_id,
			}
			r.Data = PopupResponse{
				Type:   5,
				Float:  floats,
				Data:   spin_id_data,
				UserId: 0,
			}
			return r, err
		}
		r.Msg = "no popup available"
		r.Data = PopupResponse{Type: -1, Float: floats, UserId: 0}
		return
	}

	user := u.(model.User)
	floats, err := GetFloatWindow(user, PopupTypes)

	// check redis which one has been popup
	key := "popup/records/" + time.Now().Format("2006-01-02")
	res := cache.RedisClient.HGet(context.Background(), key, strconv.FormatInt(user.ID, 10))
	if res.Err() != nil && res.Err() != redis.Nil {
		// if redis get error, return error
		fmt.Print("Redis Get failed, ", res.Err())
		return r, res.Err()
	} else if res.Err() == redis.Nil {
		// if no display record found in redis, start finding the popup window
		shouldPopupWinLose, err := model.ShouldPopupWinLose(user)
		fmt.Println("shouldPopupWinLose", shouldPopupWinLose)
		if err != nil {
			return r, err
		}
		if shouldPopupWinLose && WinLoseAvailable(PopupTypes) {
			var service WinLoseService
			data, err := service.Get(c)
			r.Data = PopupResponse{
				Type:   1,
				Float:  floats,
				Data:   data,
				UserId: user.ID,
			}
			return r, err
		}

		shouldPopupTeamUp, err := model.ShouldPopupTeamUp(user)
		fmt.Println("shouldPopupTeamUp", shouldPopupTeamUp)
		if err != nil {
			return r, err
		}
		if shouldPopupTeamUp && TeamUpAvailable(PopupTypes) {
			var service TeamUpService
			data, err := service.Get(c)
			r.Data = PopupResponse{
				Type:   data.Type,
				Float:  floats,
				Data:   data,
				UserId: user.ID,
			}
			return r, err
		}

		ShouldVIP, err := model.ShouldPopupVIP(user)
		fmt.Println("ShouldPopUpVIP", ShouldVIP)
		if err != nil {
			return r, err
		}
		if ShouldVIP && VIPAvailable(PopupTypes) {
			var service VipService
			data, err := service.Get(c)
			if err != nil {
				return r, err
			}
			r.Data = PopupResponse{
				Type:   4,
				Float:  floats,
				Data:   data,
				UserId: user.ID,
			}
			return r, err
		}

		should_spin, spin_promotion_id := SpinAvailable(PopupTypes)
		ShouldPopupSpin, err := model.ShouldPopupSpin(user, spin_promotion_id)
		if err != nil {
			return r, err
		}
		if ShouldPopupSpin && should_spin {
			spin_id_data := PopupSpinId{
				SpinId: spin_promotion_id,
			}
			r.Data = PopupResponse{
				Type:   5,
				Float:  floats,
				Data:   spin_id_data,
				UserId: user.ID,
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
			if shouldPopupWinLose && WinLoseAvailable(PopupTypes) {
				var service WinLoseService
				data, err := service.Get(c)
				r.Data = PopupResponse{
					Type:   1,
					Float:  floats,
					Data:   data,
					UserId: user.ID,
				}
				return r, err
			}
		}
		//----------------------------------------------------------------------
		if redisPopup < 3 {
			should_popup, err := model.ShouldPopupTeamUp(user)
			if err != nil {
				return r, err
			}
			if should_popup && TeamUpAvailable(PopupTypes) {
				var service TeamUpService
				data, err := service.Get(c)
				r.Data = PopupResponse{
					Type:   data.Type,
					Float:  floats,
					Data:   data,
					UserId: user.ID,
				}
				return r, err
			}
		}
		if redisPopup < 4 {
			ShouldVIP, err := model.ShouldPopupVIP(user)
			if err != nil {
				return r, err
			}
			if ShouldVIP && VIPAvailable(PopupTypes) {
				var service VipService
				data, err := service.Get(c)
				r.Data = PopupResponse{
					Type:   4,
					Float:  floats,
					Data:   data,
					UserId: user.ID,
				}
				return r, err
			}
		}
		if redisPopup < 5 {
			should_spin, spin_promotion_id := SpinAvailable(PopupTypes)
			ShouldPopupSpin, err := model.ShouldPopupSpin(user, spin_promotion_id)
			if err != nil {
				return r, err
			}
			if ShouldPopupSpin && should_spin {
				spin_id_data := PopupSpinId{
					SpinId: spin_promotion_id,
				}
				r.Data = PopupResponse{
					Type:   5,
					Float:  floats,
					Data:   spin_id_data,
					UserId: user.ID,
				}
				return r, err
			}
		}
	}
	r.Msg = "no popup available"
	r.Data = PopupResponse{Type: -1, Float: floats, UserId: user.ID}
	ShownNothing(user)
	return
}

func GetFloatWindow(user model.User, popup_types []models.Popups) (floats []PopupFloat, err error) {
	for _, popup_type := range popup_types {
		if popup_type.CanFloat {
			if popup_type.PopupType == 5 && popup_type.CanFloat {
				// spin popup float
				var spin_service SpinService
				spin_promotion_id_int, _ := strconv.Atoi(popup_type.Meta)
				if spin_service.CheckIsSpinAlive(spin_promotion_id_int) {
					// user still can spin, then we add the spin popup to float list.
					spin_promotion_id_string := strconv.Itoa(spin_promotion_id_int)
					spin, _ := model.GetSpinByPromotionId(spin_promotion_id_string)
					floats = append(floats, PopupFloat{
						Type: 5,
						Id:   spin_promotion_id_int,
						FloatUrl: serializer.Url(spin.FloatUrl),
					})
				}
			}
		}
	}
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

func TeamUpAvailable(popups []models.Popups) bool {
	for _, popup := range popups {
		if popup.PopupType == 2 || popup.PopupType == 3 {
			return true // Found a popup with PopupType == 2
		}
	}
	return false // No popup with PopupType == 2 was found
}
func VIPAvailable(popups []models.Popups) bool {
	for _, popup := range popups {
		if popup.PopupType == 4 {
			return true // Found a popup with PopupType == 3
		}
	}
	return false // No popup with PopupType == 3 was found
}
func SpinAvailable(popups []models.Popups) (bool, int) {
	for _, popup := range popups {
		if popup.PopupType == 5 && popup.CanFloat {
			var spin_service SpinService
			spin_promotion_id_int, _ := strconv.Atoi(popup.Meta)
			if spin_service.CheckIsSpinAlive(spin_promotion_id_int) {
				return true, spin_promotion_id_int // Found a popup with PopupType == 4
			}
		}
	}
	return false, 0 // No popup with PopupType == 4 was found
}

func WinLoseFloat(popups []models.Popups) bool {
	for _, popup := range popups {
		if popup.PopupType == 1 {
			return popup.CanFloat // Found a popup with PopupType == 1
		}
	}
	return false // No popup with PopupType == 1 was found
}

//	func KanDanFloat(popups []models.Popups) bool {
//	    for _, popup := range popups {
//	        if popup.PopupType == 2 {
//	            return popup.CanFloat // Found a popup with PopupType == 2
//	        }
//	    }
//	    return false // No popup with PopupType == 2 was found
//	}
func VIPFloat(popups []models.Popups) bool {
	for _, popup := range popups {
		if popup.PopupType == 4 {
			return popup.CanFloat // Found a popup with PopupType == 3
		}
	}
	return false // No popup with PopupType == 3 was found
}
func SpinFloat(popups []models.Popups) bool {
	for _, popup := range popups {
		if popup.PopupType == 5 {
			return popup.CanFloat // Found a popup with PopupType == 4
		}
	}
	return false // No popup with PopupType == 4 was found
}
func ShownNothing(user model.User) (err error) {
	key := "popup/records/" + time.Now().Format("2006-01-02")
	res := cache.RedisClient.HSet(context.Background(), key, user.ID, "6")
	if res.Err() != nil {
		fmt.Print("insert win lose popup record into redis failed ", key)
	}
	expire_time, err := strconv.Atoi(os.Getenv("POPUP_RECORD_EXPIRE_MINS"))
	cache.RedisClient.ExpireNX(context.Background(), key, time.Duration(expire_time)*time.Minute)
	return
}
