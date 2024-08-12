package serializer

import (
	"time"
	"web-api/model"

	models "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/jinzhu/copier"
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

type OutgoingTeamupHash struct {
	Id                      int64     `json:"id"`
	UserId                  int64     `json:"user_id"`
	OrderId                 string    `json:"order_id"`
	TotalAccumulatedDeposit int64     `json:"total_accumulated_deposit"`
	TotalTeamupDeposit      int64     `json:"total_teamup_deposit"`
	TotalTeamUpTarget       int64     `json:"total_teamup_target"`
	TeamupEndTime           time.Time `json:"teamup_end_time"`
	TeamupCompletedTime     time.Time `json:"teamup_completed_time"`
	Status                  int       `json:"status"`

	Nickname string            `json:"nickname"`
	Avatar   string            `json:"avatar"`
	IsParlay bool              `json:"is_parlay"`
	Bet      model.OutgoingBet `json:"bet"`
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

func BuildCustomTeamupHash(a models.Teamup, u model.User, br models.BetReport) (res OutgoingTeamupHash) {

	t := Teamup{
		Id:                      a.ID,
		UserId:                  a.UserId,
		OrderId:                 a.OrderId,
		TotalAccumulatedDeposit: a.TotalAccumulatedDeposit,
		TotalTeamupDeposit:      a.TotalTeamupDeposit,
		TotalTeamUpTarget:       a.TotalTeamUpTarget,
	}

	copier.Copy(&res, t)

	res.Nickname = u.Nickname
	res.Avatar = u.Avatar
	res.IsParlay = br.IsParlay

	return
}
