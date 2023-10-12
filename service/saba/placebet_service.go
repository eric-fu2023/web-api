package saba

import (
	"blgit.rfdev.tech/taya/game-service/saba/callback"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"time"
	"web-api/service/common"
)

type PlaceBet struct {
	Callback
	Request callback.PlaceBetRequest
}

func (c *PlaceBet) NewCallback(userId int64) {
	copier.Copy(&c.Transaction, &c.Request.Message)
	if v, e := time.Parse(time.RFC3339, c.Request.Message.BetTime); e == nil {
		c.Transaction.BetTime = v.UTC()
	}
	if v, e := time.Parse(time.RFC3339, c.Request.Message.UpdateTime); e == nil {
		c.Transaction.UpdateTime = v.UTC()
	}
	//ticketDetails := make([]ploutos.TicketDetail, 1)
	//copier.Copy(&ticketDetails[0], &c.Request.Message)
	//if v, e := time.Parse(time.RFC3339, c.Request.Message.KickOffTime); e == nil {
	//	ticketDetails[0].KickOffTime = v.UTC()
	//}
	//if v, e := time.Parse(time.RFC3339, c.Request.Message.MatchDatetime); e == nil {
	//	ticketDetails[0].MatchDatetime = v.UTC()
	//}
	ticketDetails := make([]callback.TicketDetail, 1)
	copier.Copy(&ticketDetails[0], &c.Request.Message)
	if j, e := json.Marshal(ticketDetails); e == nil {
		c.Transaction.TicketDetail = string(j)
	}
	c.Transaction.UserId = userId
	c.Transaction.ExternalUserId = c.Request.Message.UserId
	c.Transaction.BetAmount = int64(c.Request.Message.BetAmount * 100)
	c.Transaction.ActualAmount = int64(c.Request.Message.ActualAmount * 100)
	c.Transaction.CreditAmount = int64(c.Request.Message.CreditAmount * 100)
	c.Transaction.DebitAmount = int64(c.Request.Message.DebitAmount * 100)
}

func (c *PlaceBet) GetExternalUserId() string {
	return c.Request.Message.UserId
}

func PlaceBetCallback(c *gin.Context, req callback.PlaceBetRequest) (res any, err error) {
	j, _ := json.Marshal(req)
	fmt.Println("placebet: ", string(j))
	clb := PlaceBet{Request: req}
	err = common.ProcessTransaction(&clb)
	if err != nil {
		return
	}
	res = callback.PlaceBetResponse{
		BaseResponse: callback.BaseResponse{
			Status: "0",
		},
		RefId:        req.Message.RefId,
		LicenseeTxId: clb.Transaction.ID,
	}
	return
}
