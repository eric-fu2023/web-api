package service

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"
	"web-api/cache"
	"web-api/model"
	"web-api/model/avatar"
	"web-api/serializer"

	"github.com/gin-gonic/gin"
)

type WinLoseService struct {
}

type WinLosePopupResponse struct {
	CurrentRanking int               `json:"current_ranking"`
	TotalRanking   int               `json:"total_ranking"`
	GGR            float64           `json:"win_lose"`
	IsWin          bool              `json:"is_win"`
	Start          int64             `json:"start"`
	End            int64             `json:"end"`
	Member         []WinLosePopupGGR `json:"Member"`
}
type WinLosePopupGGR struct {
	GGR     float64 `json:"win_lose"`
	Ranking int     `json:"ranking"`
	Name    string  `json:"name"`
	PicSrc  string  `json:"pic_src"`
	IsMe    bool    `json:"is_me"`
}

func (service *WinLoseService) Get(c *gin.Context) (data WinLosePopupResponse, err error) {
	now := time.Now()
	yesterdayStart := time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, now.Location())
	yesterdayEnd := yesterdayStart.Add(24 * time.Hour)

	u, _ := c.Get("user")
	user := u.(model.User)
	// check if user has GGR yesterday
	type GGRRecords struct {
		GGR    float64 `json:"win_lose"`
		UserID int64   `json:"user_id"`
	}


	key := "popup/win_lose/"+now.Format("2006-01-02")
	total_ranking_key := "popup/win_lose/total_ranking/"+now.Format("2006-01-02")
	current_ranking_key := "popup/win_lose/ranking/"+now.Format("2006-01-02")
	res := cache.RedisConfigClient.HGet(context.Background(), key, strconv.FormatInt(user.ID, 10))
	GGR, err:= strconv.ParseFloat(res.Val(), 64)
	if err!=nil{
		fmt.Println("convert GGR to float64 failed!!!!", err.Error())
	}
	myGGRRecord :=GGRRecords{
		GGR: GGR,
		UserID: user.ID,
	}
	var total_ranking int
	var current_ranking int
	current_ranking_string := cache.RedisConfigClient.HGet(context.Background(), current_ranking_key,  strconv.FormatInt(user.ID, 10)).Val()
	if GGR < 0 {
		total_ranking_string := cache.RedisConfigClient.HGet(context.Background(), total_ranking_key, "lose").Val()
		total_ranking, err = strconv.Atoi(total_ranking_string)
		current_ranking, err = strconv.Atoi(current_ranking_string)
		current_ranking = -current_ranking
	}else{
		total_ranking_string := cache.RedisConfigClient.HGet(context.Background(), total_ranking_key, "win").Val()
		total_ranking, err = strconv.Atoi(total_ranking_string)
		current_ranking, err = strconv.Atoi(current_ranking_string)
	}
	var members []WinLosePopupGGR
	if GGR > 0 {
		members = append(members,
			generateMemberGGR(user.Nickname, myGGRRecord.GGR, rand.Intn(500), current_ranking, false, -1),
			generateMemberGGR(user.Nickname, myGGRRecord.GGR, 0, current_ranking, true, 0),
			generateMemberGGR(user.Nickname, myGGRRecord.GGR, -rand.Intn(500), current_ranking, false, 1))
	} else if GGR < 0 {
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
	key := "popup/records/" + time.Now().Format("2006-01-02")
	res := cache.RedisConfigClient.HSet(context.Background(), key, user.ID, "1")
	expire_time, err := strconv.Atoi(os.Getenv("POPUP_RECORD_EXPIRE_MINS"))
	cache.RedisConfigClient.ExpireNX(context.Background(), key, time.Duration(expire_time) * time.Minute)
	if res.Err() != nil {
		fmt.Print("insert win lose popup record into redis failed ", key)
	}
	return
}

func generateMemberGGR(nickname string, ggr float64, delta int, ranking int, is_me bool, index int) WinLosePopupGGR {
	var nicks []map[string]interface{}
	model.DB.Table(`ranking_nicknames`).Find(&nicks)
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
