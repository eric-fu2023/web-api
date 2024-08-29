package service

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"
	"web-api/model"
	"web-api/serializer"
	"web-api/util/i18n"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
)

type SpinService struct {
}
type SpinQueryParam struct {
	Id int `json:"id"`
}

func (service *SpinService) Get(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	spin_promotion_id := c.Query("id")
	u, _ := c.Get("user")
	user, _ := u.(model.User)
	var spin ploutos.Spins
	err = model.DB.Model(ploutos.Spins{}).Where("promotion_id = ?", spin_promotion_id).Find(&spin).Error
	if err != nil {
		r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}

	var spin_items []ploutos.SpinItem
	q := model.DB.Model(ploutos.SpinItem{}).Where("spin_id = ?", spin.ID).Order(`id DESC`)
	err = q.Find(&spin_items).Error
	if err != nil {
		r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}

	now := time.Now()
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	var spin_results []ploutos.SpinResult
	err = model.DB.Debug().Model(ploutos.SpinResult{}).Where("spin_id = ?", spin.ID).Where("user_id = ?", user.ID).Where("created_at > ?", startOfToday).Find(&spin_results).Error
	if err != nil {
		r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}
	spin_results_counts := len(spin_results)

	var data serializer.Spin
	data = serializer.BuildSpin(spin, spin_items, spin_results_counts)
	r = serializer.Response{
		Data: data,
	}
	return
}

func (service *SpinService) Result(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	spin_promotion_id := c.Query("id")
	spin_promotion_id_int, err := strconv.ParseInt(spin_promotion_id, 10, 64)
	if err != nil {
		r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}
	u, _ := c.Get("user")
	user := u.(model.User)

	// need to check if user has used all spin chances
	now := time.Now()
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	var spin ploutos.Spins
	err = model.DB.Model(ploutos.Spins{}).Where("promotion_id = ?", spin_promotion_id_int).Find(&spin).Error


	var previous_spin_result []ploutos.SpinResult
	err = model.DB.Model(ploutos.SpinResult{}).Where("user_id = ? AND spin_id = ?", user.ID, spin.ID).Where("created_at > ?", startOfToday).
		Order("created_at DESC").
		Find(&previous_spin_result).Error

	if err != nil {
		r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}
	// Check if the latest created_at time is after today
	if len(previous_spin_result) >= spin.Counts {
		r = serializer.Response{
			Code:  0,
			Msg:   "All spin chances have been used for today. Please try again tomorrow.",
			Error: "All spin chances have been used for today. Please try again tomorrow.",
			Data:  nil,
		}
		return r, nil
		// Perform your action here
	} else {
		var spinItems []ploutos.SpinItem
		q := model.DB.Model(ploutos.SpinItem{}).Where("spin_id = ?", spin.ID).Order(`id DESC`)
		err = q.Find(&spinItems).Error
		if err != nil {
			r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
			return
		}

		//------------------------------------- choose the spin item based on the probability
		var resultSpinItem ploutos.SpinItem
		var totalProbability float64
		for _, item := range spinItems {
			totalProbability += item.Probability
		}
		// Normalize the probabilities so they sum to 1.
		normalizedProbabilities := make([]float64, len(spinItems))
		for i, item := range spinItems {
			normalizedProbabilities[i] = item.Probability / totalProbability
		}
		// Generate a random number between 0 and 1.
		randomValue := rand.Float64()
		// Use the random number to select an item.
		var cumulativeProbability float64
		for i, probability := range normalizedProbabilities {
			cumulativeProbability += probability
			if randomValue < cumulativeProbability {
				resultSpinItem = spinItems[i]
				break
			}
		}
		// ------------------------------------- end of choose the spin item based on the probability

		remaining_counts := spin.Counts - len(previous_spin_result) - 1
		data := serializer.BuildSpinResult(resultSpinItem, remaining_counts)

		SpinResult := ploutos.SpinResult{
			UserID:     user.ID,
			SpinResult: data.ID,
			Redeemed:   false,
			SpinID:     spin_promotion_id_int,
		}
		err = model.DB.Create(&SpinResult).Error

		r = serializer.Response{
			Data: data,
		}
	}
	return
}

func (service *SpinService) GetHistory(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	spin_promotion_id := c.Query("id")
	spin_promotion_id_int, err := strconv.ParseInt(spin_promotion_id, 10, 64)
	u, _ := c.Get("user")
	user := u.(model.User)
	var sql_data []serializer.SpinSqlHistory
	err = model.DB.Table("spin_results").
		Joins("LEFT JOIN spin_items ON spin_items.id = spin_results.spin_result").
		Joins("LEFT JOIN spins ON spins.id = spin_results.spin_id").
		Select("spins.id as spin_id, spins.name as spin_name, spin_results.created_at, spin_results.spin_result as spin_result_id, spin_items.name as spin_result_name, spin_items.type as spin_result_type, spin_results.redeemed").
		Where("spins.promotion_id", spin_promotion_id_int).
		Where("spin_results.user_id", user.ID).
		Scan(&sql_data).Error
	var data []serializer.SpinHistory
	data = serializer.BuildSpinHistory(sql_data, i18n)
	r = serializer.Response{
		Data: data,
	}
	return

}

func (service *SpinService) GetRemainingSpinCount(user model.User, spin_id int) (remaining_counts int, err error) {
	var spin ploutos.Spins
	err = model.DB.Model(ploutos.Spins{}).Where("id = ?", spin_id).Find(&spin).Error
	if err != nil {
		return
	}
	now := time.Now()
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	var spin_results []ploutos.SpinResult
	err = model.DB.Model(ploutos.SpinResult{}).Where("spin_id = ?", spin_id).Where("user_id = ?", user.ID).Where("created_at > ?", startOfToday).Find(&spin_results).Error
	if err != nil {
		return
	}
	spin_results_counts := len(spin_results)

	return spin.Counts - spin_results_counts, err
}

func (service *SpinService) GetSpinIdFromPromotionId(spin_promotion_id int) (spin_id int, err error) {
	var spin ploutos.Spins
	err = model.DB.Debug().Model(ploutos.Spins{}).Where("promotion_id = ?", spin_promotion_id).Find(&spin).Error
	if err != nil {
		fmt.Println("get spin id error", err)
	}
	fmt.Println("get spin ", spin.ID)
	return int(spin.ID), err
}

func (service *SpinService) CheckIsSpinAlive(spin_promotion_id int) (isAlive bool) {
	var promotion ploutos.Promotion
	err := model.DB.Where("id", spin_promotion_id).First(&promotion).Error
	if err!= nil{
		fmt.Println("get spin promotion error", err)
	}
	if promotion.EndAt.After(time.Now()) && promotion.StartAt.Before(time.Now()){
		return true
	}
	return false
}

