package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"
	"web-api/cache"
	"web-api/model"
	"web-api/model/avatar"
	"web-api/serializer"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm/logger"
)

type TeamUpService struct {
}

type TeamUpPopupResponse struct {
	Id      int64  `json:"id"`
	OrderId string `json:"order_id"`
	Status  int    `json:"status"`
	// TotalTeamupDeposit int64                   `json:"total_deposit"`
	TotalTeamUpTarget float64                 `json:"total_target"`
	// Remaining         int64                   `json:"remaining"`
	// Saved             int64                   `json:"saved"`
	Progress          float64                 `json:"progress"`
	Start             int64                   `json:"start"`
	End               int64                   `json:"end"`
	Type              int                     `json:"type"`
	Members           []TeamUpPopupMemberInfo `json:"members"`
}
type TeamUpPopupMemberInfo struct {
	TotalTeamUpTarget float64 `json:"total_target"`
	Ranking           int64   `json:"ranking"`
	Name              string  `json:"name"`
	PicSrc            string  `json:"pic_src"`
	IsMe              bool    `json:"is_me"`
}

func (service *TeamUpService) Get(c *gin.Context) (data TeamUpPopupResponse, err error) {
	now := time.Now()
	TodayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yesterdayStart := time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, now.Location())
	yesterdayEnd := yesterdayStart.Add(24 * time.Hour)

	u, _ := c.Get("user")
	user := u.(model.User)

	var team_up ploutos.Teamup
	// status = 1 is success,    status = 0 is onging
	err = model.DB.Model(ploutos.Teamup{}).Where("user_id = ? AND updated_at < ? AND updated_at > ? AND status in (1,0) AND total_fake_progress !=0", user.ID, TodayStart, yesterdayStart).Order("status DESC, total_fake_progress DESC").First(&team_up).Error
	if errors.Is(err, logger.ErrRecordNotFound) {
		err = nil
		// if no team up record, we return nil
		return TeamUpPopupResponse{}, err
	}
	if err != nil {
		fmt.Println("Get teamup err", err.Error())
		return TeamUpPopupResponse{}, err
	}

	var teamup_type int
	members := make([]TeamUpPopupMemberInfo, 0)
	if team_up.Status == 0 {
		// ongoing
		teamup_type = 3
	} else {
		// success
		teamup_type = 2
		members = GenerateMembersForTeamUpSuccess(user, team_up.TotalTeamUpTarget)
	}
	// follow 2 fields are only used in batace due to INR need to rounding
	// saved:=int64(float64(team_up.TotalTeamUpTarget) / 100 * float64(team_up.TotalFakeProgress)/10000)
	// remaining:=(team_up.TotalTeamUpTarget-saved) / 100
	data = TeamUpPopupResponse{
		Id:      team_up.ID,
		OrderId: team_up.OrderId,
		Status:  team_up.Status,
		// TotalTeamupDeposit: team_up.TotalTeamupDeposit / 100,
		TotalTeamUpTarget: float64(team_up.TotalTeamUpTarget) / 100,
		Progress:          float64(team_up.TotalFakeProgress) / 100,
		// Remaining:         remaining,
		// Saved:             saved,
		Start:             yesterdayStart.Unix(),
		End:               yesterdayEnd.Unix(),
		Type:              teamup_type,
		Members:           members,
	}
	service.Shown(c)
	return data, nil
}

func (service *TeamUpService) Shown(c *gin.Context) (r serializer.Response, err error) {
	u, _ := c.Get("user")
	user := u.(model.User)
	if err != nil {
		return
	}
	key := "popup/records/" + time.Now().Format("2006-01-02")
	res := cache.RedisClient.HSet(context.Background(), key, user.ID, "3")
	expire_time, err := strconv.Atoi(os.Getenv("POPUP_RECORD_EXPIRE_MINS"))
	cache.RedisClient.ExpireNX(context.Background(), key, time.Duration(expire_time)*time.Minute)
	if res.Err() != nil {
		fmt.Print("insert win lose popup record into redis failed ", key)
	}
	return
}

func GenerateMembersForTeamUpSuccess(user model.User, total_team_up_target int64) []TeamUpPopupMemberInfo {
	var nicks []map[string]interface{}
	model.DB.Table(`ranking_nicknames`).Find(&nicks)
	var ranking_higher_user_nickname string
	var ranking_lower_user_nickname string
	if len(nicks) > 0 {
		rand.Seed(time.Now().UnixNano())
		r1 := rand.Intn(len(nicks))
		r2 := rand.Intn(len(nicks))
		ranking_higher_user_nickname = nicks[r1]["first_name"].(string) + " " + nicks[r2]["last_name"].(string)
		r3 := rand.Intn(len(nicks))
		r4 := rand.Intn(len(nicks))
		ranking_lower_user_nickname = nicks[r3]["first_name"].(string) + " " + nicks[r4]["last_name"].(string)
	}
	resp := make([]TeamUpPopupMemberInfo, 0)

	team_up_ranking_param_a, err := strconv.ParseInt(os.Getenv("TEAMUP_RANKING_PARAM_A"), 10, 64)
	if err != nil {
		fmt.Println("There is a error in strconv for min team up value, TEAMUP_RANKING_PARAM_A")
	}
	team_up_ranking_param_b, err := strconv.ParseInt(os.Getenv("TEAMUP_RANKING_PARAM_B"), 10, 64)
	if err != nil {
		fmt.Println("There is a error in strconv for min team up value, TEAMUP_RANKING_PARAM_B")
	}

	estimated_ranking := team_up_ranking_param_a / (team_up_ranking_param_b * total_team_up_target)
	if estimated_ranking < 2 {
		estimated_ranking = 2
	}
	resp = append(resp, TeamUpPopupMemberInfo{
		TotalTeamUpTarget: float64(total_team_up_target)/100 + float64(rand.Int63n(50)),
		Ranking:           estimated_ranking - 1,
		Name:              ranking_higher_user_nickname,
		PicSrc:            avatar.GetRandomAvatarUrl(),
		IsMe:              false,
	})
	resp = append(resp, TeamUpPopupMemberInfo{
		TotalTeamUpTarget: float64(total_team_up_target) / 100,
		Ranking:           estimated_ranking,
		Name:              user.Nickname,
		PicSrc:            serializer.Url(user.Avatar),
		IsMe:              true,
	})

	ranking_lower_total_target := total_team_up_target - rand.Int63n(50)
	team_up_min, err := strconv.ParseInt(os.Getenv("TEAMUP_MIN"), 10, 64)
	if err != nil {
		fmt.Println("There is a error in strconv for min team up value, TEAMUP_MIN")
	}
	if ranking_lower_total_target < team_up_min {
		ranking_lower_total_target = team_up_min
	}

	resp = append(resp, TeamUpPopupMemberInfo{
		TotalTeamUpTarget: float64(ranking_lower_total_target) / 100,
		Ranking:           estimated_ranking + 1,
		Name:              ranking_lower_user_nickname,
		PicSrc:            avatar.GetRandomAvatarUrl(),
		IsMe:              false,
	})
	return resp
}
