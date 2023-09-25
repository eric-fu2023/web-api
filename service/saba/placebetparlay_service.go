package saba

import (
	"blgit.rfdev.tech/taya/game-service/saba/callback"
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
	if j, e := json.Marshal(c.Request.Message.TicketDetail); e == nil {
		c.Transaction.TicketDetail = string(j)
	}
	for _, t := range c.Request.Message.Txns {
		if j, e := json.Marshal(t); e == nil {
			c.Transaction.RefId = t.RefId
			c.Transaction.BetAmount = int64(t.BetAmount * 100)
			c.Transaction.ActualAmount = c.Transaction.BetAmount
			c.Transaction.CreditAmount = int64(t.CreditAmount * 100)
			c.Transaction.DebitAmount = int64(t.DebitAmount * 100)
			c.Transaction.ParlayDetail = string(j)
		}
	}
	c.Transaction.UserId = userId
	c.Transaction.ExternalUserId = c.Request.Message.UserId
}

func (c *PlaceBetParlay) GetExternalUserId() string {
	return c.Request.Message.UserId
}

func PlaceBetParlayCallback(c *gin.Context, req callback.PlaceBetParlayRequest) (res any, err error) {
	j, _ := json.Marshal(req)
	fmt.Println("placebetparlay: ", string(j))
	var r []callback.PlaceBetParlayResponseTxns
	for _, t := range req.Message.Txns {
		var newReq callback.PlaceBetParlayRequest
		copier.Copy(&newReq, req)
		newReq.Message.Txns = []callback.PlaceBetParlayTxns{t}
		clb := PlaceBetParlay{Request: newReq}
		err = service.ProcessTransaction(&clb)
		if err != nil {
			return
		}
		r = append(r, callback.PlaceBetParlayResponseTxns{
			RefId:        t.RefId,
			LicenseeTxId: clb.Transaction.ID,
		})
	}
	res = callback.PlaceBetParlayResponse{
		BaseResponse: callback.BaseResponse{
			Status: "0",
		},
		Txns: r,
	}
	return
}
