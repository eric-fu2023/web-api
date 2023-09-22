package saba

import (
	"blgit.rfdev.tech/taya/game-service/saba/callback"
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"time"
	"web-api/service"
)

type PlaceBetParlay struct {
	Callback
	Request callback.PlaceBetParlayRequest
}

func (c *PlaceBetParlay) NewCallback(userId int64) {
	copier.Copy(&c.Transaction, &c.Request.Message)
	if v, e := time.Parse(time.RFC3339, c.Request.Message.BetTime); e == nil {
		c.Transaction.BetTime = v.UTC()
	}
	if v, e := time.Parse(time.RFC3339, c.Request.Message.UpdateTime); e == nil {
		c.Transaction.UpdateTime = v.UTC()
	}
	ticketDetails := make([]ploutos.TicketDetail, len(c.Request.Message.TicketDetail))
	for i, t := range c.Request.Message.TicketDetail {
		copier.Copy(&ticketDetails[i], &t)
		if v, e := time.Parse(time.RFC3339, t.KickOffTime); e == nil {
			ticketDetails[i].KickOffTime = v.UTC()
		}
		if v, e := time.Parse(time.RFC3339, t.MatchDatetime); e == nil {
			ticketDetails[i].MatchDatetime = v.UTC()
		}
	}
	if j, e := json.Marshal(ticketDetails); e == nil {
		c.Transaction.TicketDetail = string(j)
	}
	if j, e := json.Marshal(c.Request.Message.Txns); e == nil {
		c.Transaction.ParlayDetail = string(j)
	}
	c.Transaction.UserId = userId
	c.Transaction.ExternalUserId = c.Request.Message.UserId
	c.Transaction.BetAmount = int64(c.Request.Message.TotalBetAmount * 100)
	c.Transaction.ActualAmount = c.Transaction.BetAmount
	c.Transaction.CreditAmount = int64(c.Request.Message.CreditAmount * 100)
	c.Transaction.DebitAmount = int64(c.Request.Message.DebitAmount * 100)
}

func (c *PlaceBetParlay) GetExternalUserId() string {
	return c.Request.Message.UserId
}

func PlaceBetParlayCallback(c *gin.Context, req callback.PlaceBetParlayRequest) (res any, err error) {
	j, _ := json.Marshal(req)
	fmt.Println("placebetparlay: ", string(j))
	clb := PlaceBetParlay{Request: req}
	err = service.ProcessTransaction(&clb)
	if err != nil {
		return
	}
	res = callback.PlaceBetParlayResponse{
		BaseResponse: callback.BaseResponse{
			Status: "0",
		},
		//Txns: callback.PlaceBetParlayResponseTxns{
		//	RefId:        req.Message.,
		//	LicenseeTxId: clb.Transaction.ID,
		//},
	}
	return
}
