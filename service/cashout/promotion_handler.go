package cashout

import (
	"context"
	"web-api/model"
	"web-api/service"
)

func HandlePromotion(c context.Context, order model.CashOrder) {
	service.HandleCashMethodPromotion(c, order)
}
