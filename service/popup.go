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
type PopupSpinId struct {
	SpinId int `json:"spin_id"`
}
type PopupResponse struct {
	Type  int          `json:"type"`
	Float []PopupFloat `json:"float"`
	Data  interface{}  `json:"data"`
}

type PopupFloat struct {
	Type int64 `json:"type"`
	Id   int   `json:"id"`
}

func (service *PopupService) ShowPopup(c *gin.Context) (r serializer.Response, err error) {
	// check what to popup
	PopupTypes, err := model.GetPopupList(service.Condition)

	u, isUser := c.Get("user")
	if !isUser {
		// if not login, show spin only
		should_spin, spin_promotion_id := SpinAvailable(PopupTypes)
		spin_promotion_id_int, _ := strconv.Atoi(spin_promotion_id)
		var spin_service SpinService
		spin_id_int, _ := spin_service.GetSpinIdFromPromotionId(spin_promotion_id_int)

		var floats []PopupFloat

		if should_spin {
			floats = append(floats, PopupFloat{
				Type: 5,
				Id:   spin_id_int,
			})
			spin_id_data := PopupSpinId{
				SpinId: spin_id_int,
			}
			r.Data = PopupResponse{
				Type:  5,
				Float: floats,
				Data:  spin_id_data,
			}
			return r, err
		}
		r.Msg = "no popup available"
		r.Data = PopupResponse{Type: -1, Float: floats}
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
				Type:  1,
				Float: floats,
				Data:  data,
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
				Type:  data.Type,
				Float: floats,
				Data:  data,
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
				Type:  4,
				Float: floats,
				Data:  data,
			}
			return r, err
		}

		should_spin, spin_promotion_id := SpinAvailable(PopupTypes)
		spin_promotion_id_int, _ := strconv.Atoi(spin_promotion_id)
		var spin_service SpinService
		spin_id_int, _ := spin_service.GetSpinIdFromPromotionId(spin_promotion_id_int)
		ShouldPopupSpin, err := model.ShouldPopupSpin(user, spin_id_int)
		fmt.Println("ShouldPopupSpin", ShouldPopupSpin)
		fmt.Println("ShouldPopupSpin", should_spin)
		if err != nil {
			return r, err
		}
		if ShouldPopupSpin && should_spin {
			spin_id_data := PopupSpinId{
				SpinId: spin_id_int,
			}
			r.Data = PopupResponse{
				Type:  5,
				Float: floats,
				Data:  spin_id_data,
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
					Type:  1,
					Float: floats,
					Data:  data,
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
					Type:  data.Type,
					Float: floats,
					Data:  data,
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
					Type:  4,
					Float: floats,
					Data:  data,
				}
				return r, err
			}
		}
		if redisPopup < 5 {
			should_spin, spin_promotion_id := SpinAvailable(PopupTypes)
			spin_promotion_id_int, _ := strconv.Atoi(spin_promotion_id)
			var spin_service SpinService
			spin_id_int, _ := spin_service.GetSpinIdFromPromotionId(spin_promotion_id_int)
			ShouldPopupSpin, err := model.ShouldPopupSpin(user, spin_id_int)
			if err != nil {
				return r, err
			}
			if ShouldPopupSpin && should_spin {
				spin_id_data := PopupSpinId{
					SpinId: spin_id_int,
				}
				r.Data = PopupResponse{
					Type:  5,
					Float: floats,
					Data:  spin_id_data,
				}
				return r, err
			}
		}
	}
	r.Msg = "no popup available"
	r.Data = PopupResponse{Type: -1, Float: floats}
	return
}

func GetFloatWindow(user model.User, popup_types []models.Popups) (floats []PopupFloat, err error) {
	for _, popup_type := range popup_types {
		if popup_type.CanFloat {
			if popup_type.PopupType == 5 {
				// spin popup float
				var service SpinService
				spin_promotion_id_int, err := strconv.Atoi(popup_type.Meta)
				spin_id_int, err := service.GetSpinIdFromPromotionId(spin_promotion_id_int)
				// check if user still has spin chances
				if err != nil {
					fmt.Println("GetSpinIdFromPromotionId err")
					return nil, err
				}
					// user still can spin, then we add the spin popup to float list.
					floats = append(floats, PopupFloat{
						Type: 5,
						Id:   spin_id_int,
					})
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
func SpinAvailable(popups []models.Popups) (bool, string) {
	for _, popup := range popups {
		if popup.PopupType == 5 && popup.CanFloat{
			return true, popup.Meta // Found a popup with PopupType == 4
		}
	}
	return false, "" // No popup with PopupType == 4 was found
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