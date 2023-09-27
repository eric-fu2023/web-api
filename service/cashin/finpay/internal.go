package cashin_finpay

import (
	"web-api/model"
	"web-api/serializer"
	"web-api/service/cashin"
	"web-api/util"

	"github.com/gin-gonic/gin"
)

type ManualCloseService struct {
	OrderNumber    string `json:"order_number" form:"order_number" binding:"required"`
	ManualClosedBy int64  `json:"manual_closed_by" form:"manual_closed_by" binding:"required"`
	Remark         string `json:"remark" form:"remark"`
}

func (s ManualCloseService) Do(c *gin.Context) (r serializer.Response, err error) {
	if _, err = cashin.CloseCashInOrder(c, s.OrderNumber, 0, 0, 0, util.JSON(s), "", s.Remark, model.DB, s.ManualClosedBy); err != nil {
		r = serializer.Err(c, s, serializer.CodeGeneralError, "", err)
		return
	}
	return
}
