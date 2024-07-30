package service

import (
	"math/rand"
	"time"
	"web-api/model"
	"web-api/serializer"
	"web-api/util/i18n"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
)

type SpinService struct {
}

func (service *SpinService) Get(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	var spinItems []ploutos.SpinItem
	q := model.DB.Model(ploutos.SpinItem{}).Order(`id DESC`)
	err = q.Find(&spinItems).Error
	if err != nil {
		r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}
	data := make([]serializer.SpinItem, 0)
	for _, spinItem := range spinItems {
		data = append(data, serializer.BuildSpinItem(spinItem))
	}
	r = serializer.Response{
		Data: data,
	}
	return
}

func (service *SpinService) Result(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")
	user := u.(model.User)

	var previous_spin_result ploutos.SpinResult
	err = model.DB.Model(ploutos.SpinResult{}).Where("user_id = ? AND deleted_at IS NULL", user.ID).
		Order("created_at DESC").
		First(&previous_spin_result).Error

	now := time.Now()
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// Check if the latest created_at time is after today
	if previous_spin_result.CreatedAt.After(startOfToday){
		r = serializer.Response{
			Code:  0,
			Msg:   "All spin chances have been used for today. Please try again tomorrow.",
			Error: "All spin chances have been used for today. Please try again tomorrow.",
			Data: nil,
		}
		return r, nil
		// Perform your action here
	} else {

		var spinItems []ploutos.SpinItem
		q := model.DB.Model(ploutos.SpinItem{}).Order(`id DESC`)
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

		data := serializer.BuildSpinResult(resultSpinItem)

		SpinResult := ploutos.SpinResult{
			UserID:     user.ID,
			SpinResult: data.ID,
			Redeemed:   false,
		}
		err = model.DB.Create(&SpinResult).Error

		r = serializer.Response{
			Data: data,
		}
	}
	return
}
