package serializer

import (
	"strconv"
	"web-api/conf/consts"
	"web-api/util"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
)

type BetReport struct {
	OrderId    string   `json:"order_id"`
	BusinessId string   `json:"business_id"`
	Ts         int64    `json:"ts"`
	SettleTs   int64    `json:"settle_ts"`
	Status     int64    `json:"status"`
	IsParlay   bool     `json:"is_parlay"`
	MatchCount int64    `json:"match_count"`
	BetType    string   `json:"bet_type"`
	Stake      float64  `json:"stake"`
	MaxReturn  float64  `json:"max_return,omitempty"`
	Won        *float64 `json:"won,omitempty"`
	Outcome    *string  `json:"outcome,omitempty"` // only for imsb
	Bets       []Bet    `json:"bets,omitempty"`
	Game       *Game    `json:"game,omitempty"`
	Voucher    *Voucher `json:"voucher,omitempty"`
}

func BuildBetReport(c *gin.Context, a ploutos.BetReport) (b BetReport) {
	deviceInfo, _ := util.GetDeviceInfo(c)
	b = BetReport{
		OrderId:    a.OrderId,
		BusinessId: a.BusinessId,
		Ts:         a.BetTime.Unix(),
		Status:     a.Status,
		IsParlay:   a.IsParlay,
		MatchCount: a.MatchCount,
		BetType:    a.BetType,
		Stake:      float64(a.Bet) / 100,
	}
	if a.RewardTime != nil {
		b.SettleTs = a.RewardTime.Unix()
	}
	if a.MaxWinAmount != "" {
		if v, e := strconv.ParseFloat(a.MaxWinAmount, 64); e == nil {
			b.MaxReturn = v / 100
		}
	}
	if a.Status == 5 {
		t := float64(a.Win) / 100
		b.Won = &t
	}
	if a.GameType == consts.GameVendor["imsb"] && a.Provider != "" { // only for imsb
		b.Outcome = &a.Provider
	}
	if len(a.Bets) > 0 {
		for _, l := range a.Bets {
			b.Bets = append(b.Bets, BuildBet(c, l))
		}
	}
	if a.Game != nil {
		t := BuildGame(c, *a.Game)
		b.Game = &t
	}
	if a.Voucher.ID != 0 {
		t := BuildVoucher(a.Voucher, deviceInfo.Platform)
		b.Voucher = &t
	} else {
		if len(a.ImVoucher) > 0 {
			t := BuildVoucher(a.ImVoucher[0], deviceInfo.Platform)
			b.Voucher = &t
		}
	}
	return
}

type PaginatedBetReport struct {
	TotalCount  int64       `json:"total_count"`
	TotalAmount float64     `json:"total_amount"`
	TotalWin    float64     `json:"total_win"`
	List        []BetReport `json:"list,omitempty"`
}

func BuildPaginatedBetReport(c *gin.Context, a []ploutos.BetReport, total, amount, win int64) (b PaginatedBetReport) {
	b = PaginatedBetReport{
		TotalCount:  total,
		TotalAmount: float64(amount) / 100,
		TotalWin:    float64(win) / 100,
	}
	for _, aa := range a {
		b.List = append(b.List, BuildBetReport(c, aa))
	}
	return
}
