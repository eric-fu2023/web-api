package cashin_finpay

import (
	"web-api/model"
	"web-api/serializer"
	"web-api/service/cashin"
	"web-api/util"

	"github.com/gin-gonic/gin"
)

type ManualCloseService struct {
	OrderNumber           string `json:"order_number" form:"order_number" binding:"required"`
	ActualAmount          int64  `json:"actual_amount" form:"actual_amount"`
	BonusAmount           int64  `json:"bonus_amount" form:"bonus_amount"`
	AdditionalWagerChange int64  `json:"additional_wager_change" form:"additional_wager_change"`
	TransactionType       int64  `json:"transaction_type," form:"transaction_type"`
}

func (s ManualCloseService) Do(c *gin.Context) (r serializer.Response, err error) {
	if s.TransactionType == 0 {
		s.TransactionType = 10000
	}
	if _, err = cashin.CloseCashInOrder(c, s.OrderNumber, s.ActualAmount, s.BonusAmount, s.AdditionalWagerChange, util.JSON(s), model.DB, s.TransactionType); err != nil {
		r = serializer.Err(c, s, serializer.CodeGeneralError, "", err)
		return
	}
	return
}
