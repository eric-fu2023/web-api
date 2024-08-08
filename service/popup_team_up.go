package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
	"web-api/cache"
	"web-api/model"
	"web-api/serializer"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm/logger"
)

type TeamUpService struct {
}

type TeamUpPopupResponse struct {
	OrderId string               `json:"order_id"`
	Status   int               `json:"status"`
	TotalTeamupDeposit            int64           `json:"total_deposit"`
	TotalTeamUpTarget          int64              `json:"total_target"`
	Percent         int64 `json:"percent"`
	Start          int64             `json:"start"`
	End            int64             `json:"end"`
	Type int `json:"type"`
}

func (service *TeamUpService) Get(c *gin.Context) (data TeamUpPopupResponse, err error) {
	now := time.Now()
	TodayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yesterdayStart := time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, now.Location())
	yesterdayEnd := yesterdayStart.Add(24 * time.Hour)

	u, _ := c.Get("user")
	user := u.(model.User)

	var team_up ploutos.Teamup
	// status = 2 is success,    status = 0 is onging
	err = model.DB.Model(ploutos.Teamup{}).Where("user_id = ? AND created_at < ? AND created_at > ? AND status in (2,0)", user.ID, TodayStart, yesterdayStart).Order("status DESC, total_teamup_deposit DESC").First(&team_up).Error
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
		if team_up.Status == 0 {
			teamup_type=3
		} else {
			teamup_type=2
		}
		data = TeamUpPopupResponse{
			OrderId: team_up.OrderId,
			Status:team_up.Status,
			TotalTeamupDeposit:team_up.TotalTeamupDeposit,
			TotalTeamUpTarget :team_up.TotalTeamUpTarget,
			Percent:team_up.TotalTeamupDeposit*100/team_up.TotalTeamUpTarget,
			Start:          yesterdayStart.Unix(),
			End:            yesterdayEnd.Unix(),
			Type: teamup_type,
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
	cache.RedisClient.ExpireNX(context.Background(), key, time.Duration(expire_time) * time.Minute)
	if res.Err() != nil {
		fmt.Print("insert win lose popup record into redis failed ", key)
	}
	return
}
