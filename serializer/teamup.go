package serializer

import (
	models "blgit.rfdev.tech/taya/ploutos-object"
)

type Teamup struct {
	Id                      int64 `json:"id"`
	UserId                  int64 `json:"user_id"`
	OrderId                 int64 `json:"order_id"`
	TotalAccumulatedDeposit int64 `json:"total_accumulated_deposit"`
	TotalRequiredDeposit    int64 `json:"total_required_deposit"`
	Status                  int   `json:"status"`
}

func BuildTeamup(a models.Teamup) (res Teamup) {

	res = Teamup{
		Id:                      a.ID,
		UserId:                  a.UserId,
		OrderId:                 a.OrderId,
		TotalAccumulatedDeposit: a.TotalAccumulatedDeposit,
		TotalRequiredDeposit:    a.TotalRequiredDeposit,
	}

	return
}
