package serializer

import (
	"time"

	models "blgit.rfdev.tech/taya/ploutos-object"
)

type Teamup struct {
	Id                      int64     `json:"id"`
	UserId                  int64     `json:"user_id"`
	OrderId                 string    `json:"order_id"`
	TotalAccumulatedDeposit int64     `json:"total_accumulated_deposit"`
	TotalTeamupDeposit      int64     `json:"total_teamup_deposit"`
	TotalTeamUpTarget       int64     `json:"total_teamup_target"`
	TeamupEndTime           time.Time `json:"teamup_end_time"`
	TeamupCompletedTime     time.Time `json:"teamup_completed_time"`
	Status                  int       `json:"status"`
}

type TeamupEntry struct {
	TeamupId                 int64     `json:"teamup_id"`
	UserId                   int64     `json:"user_id"`
	ContributedTeamupDeposit int64     `json:"contributed_teamup_deposit"`
	ContributedTeamupTargete int64     `json:"contributed_teamup_target"`
	TeamupEndTime            time.Time `json:"teamup_end_time"`
	TeamupCompletedTime      time.Time `json:"teamup_completed_time"`
}

func BuildTeamup(a models.Teamup) (res Teamup) {

	res = Teamup{
		Id:                      a.ID,
		UserId:                  a.UserId,
		OrderId:                 a.OrderId,
		TotalAccumulatedDeposit: a.TotalAccumulatedDeposit,
		TotalTeamupDeposit:      a.TotalTeamupDeposit,
		TotalTeamUpTarget:       a.TotalTeamUpTarget,
	}

	return
}
