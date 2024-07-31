package service

import (
	"context"
	"fmt"
	"math/rand"
	"time"
	"web-api/cache"
	"web-api/model"
	"web-api/model/avatar"
	"web-api/serializer"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
)

type WinLoseService struct {
}

type WinLosePopupResponse struct {
	CurrentRanking int               `json:"current_ranking"`
	TotalRanking   int               `json:"total_ranking"`
	GGR            float64           `json:"ggr"`
	IsWin          bool              `json:"is_win"`
	Start          int64             `json:"start"`
	End            int64             `json:"end"`
	Member         []WinLosePopupGGR `json:"Member"`
}
type WinLosePopupGGR struct {
	GGR     float64 `json:"ggr"`
	Ranking int     `json:"ranking"`
	Name    string  `json:"name"`
	PicSrc  string  `json:"pic_src"`
	IsMe    bool    `json:"is_me"`
}

func (service *WinLoseService) Get(c *gin.Context) (data WinLosePopupResponse, err error) {
	settleStatus := []int64{5}
	now := time.Now()
	yesterdayStart := time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, now.Location())
	yesterdayEnd := yesterdayStart.Add(24 * time.Hour)

	u, _ := c.Get("user")
	user := u.(model.User)
	// check if user has GGR yesterday
	type GGRRecords struct {
		GGR    float64 `json:"ggr"`
		UserID int64   `json:"user_id"`
	}
	var ggrRecords []GGRRecords
	err = model.DB.Model(ploutos.BetReport{}).Where("status = ? AND bet_time BETWEEN ? AND ? ", settleStatus, yesterdayStart, yesterdayEnd).Select("user_id, SUM(win - bet) as ggr").Group("user_id").Order("ggr desc").Find(&ggrRecords).Error
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	var myGGRRecord GGRRecords
	for _, record := range ggrRecords {
		if record.UserID == user.ID {
			myGGRRecord = record
			break
		}
	}
	total_ranking := len(ggrRecords)
	current_ranking := 0
	min := 80
	max := 200
	if myGGRRecord.GGR > 0 {
		for index, record := range ggrRecords {
			if record.GGR < 0 {
				total_ranking = index - 1
				break
			}
			if record.UserID == user.ID {
				current_ranking = index + 1
			}
		}
		random_multiplier := rand.Intn(max-min+1) + min
		current_ranking = current_ranking * random_multiplier
		total_ranking = (total_ranking + 1) * random_multiplier
	} else if myGGRRecord.GGR < 0 {
		for index, record := range ggrRecords {
			if record.GGR > 0 {
				total_ranking = total_ranking - 1
			}
			if record.UserID == user.ID {
				current_ranking = len(ggrRecords) - index
				break
			}
		}
		random_multiplier := rand.Intn(max-min+1) + min
		current_ranking = current_ranking * random_multiplier
		total_ranking = (total_ranking + 1) * random_multiplier
	}
	var members []WinLosePopupGGR
	if myGGRRecord.GGR > 0 {
		members = append(members,
			generateMemberGGR(user.Nickname, myGGRRecord.GGR, rand.Intn(500), current_ranking, false, -1),
			generateMemberGGR(user.Nickname, myGGRRecord.GGR, 0, current_ranking, true, 0),
			generateMemberGGR(user.Nickname, myGGRRecord.GGR, -rand.Intn(500), current_ranking, false, 1))
	} else if myGGRRecord.GGR < 0 {
		members = append(members,
			generateMemberGGR(user.Nickname, myGGRRecord.GGR, rand.Intn(500)+500, current_ranking, false, 2),
			generateMemberGGR(user.Nickname, myGGRRecord.GGR, rand.Intn(500), current_ranking, false, 1),
			generateMemberGGR(user.Nickname, myGGRRecord.GGR, 0, current_ranking, true, 0))
	}
	data = WinLosePopupResponse{
		CurrentRanking: current_ranking,
		TotalRanking:   total_ranking,
		Start:          yesterdayStart.Unix(),
		End:            yesterdayEnd.Unix(),
		GGR:            myGGRRecord.GGR / 100,
		IsWin:          myGGRRecord.GGR > 0,
		Member:         members,
	}
	_, shown_err := service.Shown(c)
	if shown_err != nil {
		return WinLosePopupResponse{}, shown_err
	}
	return
}

func (service *WinLoseService) Shown(c *gin.Context) (r serializer.Response, err error) {
	u, _ := c.Get("user")
	user := u.(model.User)
	if err != nil {
		return
	}
	PopupRecord := ploutos.PopupRecord{
		UserID: user.ID,
		Type:   1,
	}
	err = model.DB.Create(&PopupRecord).Error

	var popup_service PopupService
	key := popup_service.buildKey(user.ID)
	res := cache.RedisClient.Set(context.Background(), key, "1", time.Hour*24)
	if res.Err() != nil {
		fmt.Print("insert win lose popup record into redis failed ", key)
	}
	return
}

func generateMemberGGR(nickname string, ggr float64, delta int, ranking int, is_me bool, index int) WinLosePopupGGR {
	var nicks []map[string]interface{}
	model.DB.Table(`nicknames`).Find(&nicks)
	var name string
	if len(nicks) > 0 {
		rand.Seed(time.Now().UnixNano())
		r1 := rand.Intn(len(nicks))
		r2 := rand.Intn(len(nicks))
		name = nicks[r1]["first_name"].(string) + " " + nicks[r2]["last_name"].(string)
	}
	if ggr < 0 {
		index = index * -1
		delta = delta * -1
	}
	if is_me {
		name = nickname
	}
	resp := WinLosePopupGGR{
		GGR:     (ggr + float64(delta)) / 100.0,
		Ranking: ranking + index,
		Name:    name,
		PicSrc:  avatar.GetRandomAvatarUrl(),
		IsMe:    is_me,
	}
	return resp
}
