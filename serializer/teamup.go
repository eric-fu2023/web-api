package serializer

import (
	"math/rand"
	"sort"
	"time"
	"web-api/model"
	"web-api/model/avatar"

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
	Id                      int64  `json:"id"`
	UserId                  int64  `json:"user_id"`
	OrderId                 string `json:"order_id"`
	TotalAccumulatedDeposit int64  `json:"total_accumulated_deposit"`
	TotalTeamupDeposit      int64  `json:"total_teamup_deposit"`
	TotalTeamUpTarget       int64  `json:"total_teamup_target"`
	TeamupEndTime           int64  `json:"teamup_end_time"`
	TeamupCompletedTime     int64  `json:"teamup_completed_time"`
	Status                  int    `json:"status"`

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

type OtherTeamupContribution struct {
	Nickname string  `json:"nickname"`
	Time     int64   `json:"time"`
	Amount   float64 `json:"amount"`
	Avatar   string  `json:"avatar"`
	IsReal   bool    `json:"is_real"`
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

func BuildCustomTeamupHash(a models.Teamup, u model.User, isParlay bool) (res OutgoingTeamupHash) {

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
	res.IsParlay = isParlay

	return
}

func BuildCustomTeamupGameHash(teamupId int64, u model.User) (res OutgoingTeamupHash) {

	res.Id = teamupId
	res.Nickname = u.Nickname
	res.Avatar = u.Avatar

	return
}

func GenerateOtherTeamups(nicknames []string, successTeamups model.TeamupSuccess) (res []OtherTeamupContribution) {

	for i := 0; i < len(nicknames); i++ {
		item := OtherTeamupContribution{
			Nickname: nicknames[i],
			Time:     time.Now().UTC().Unix() - (int64(rand.Intn(1799)) + 1),
			Amount:   50 + float64(rand.Intn(449)+1),
			Avatar:   avatar.GetRandomAvatarUrlForTeamup(),
			IsReal:   false,
		}
		res = append(res, item)
	}

	trueCount := 0

	for i := 0; i < len(successTeamups); i++ {
		item := OtherTeamupContribution{
			Nickname: successTeamups[i].Nickname,
			Time:     successTeamups[i].Time,
			Amount:   float64(successTeamups[i].Amount) / 100,
			Avatar:   Url(successTeamups[i].Avatar),
			IsReal:   true,
		}
		trueCount++
		res = append(res, item)
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].Time > res[j].Time
	})

	return
}

