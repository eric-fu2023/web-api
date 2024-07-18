package service

import (
	"errors"
	"log"
	"strconv"
	"time"
	"web-api/model"
	"web-api/serializer"
	"web-api/service/common"
	"web-api/util"
	"web-api/util/i18n"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type GiftSendRequestService struct {
	GiftId       int64  `form:"gift_id" json:"gift_id"`
	Quantity     int    `form:"quantity" json:"quantity"`
	LiveStreamId int64  `form:"live_stream_id" json:"live_stream_id"`
	VipId        int64  `form:"vip_id" json:"vip_id"`
	StreamerName string `form:"streamer_name" json:"streamer_name"`
	AvatarUrl    string `form:"avatar_url" json:"avatar_url"`
}

type GiftRecordListService struct {
	Start string `form:"start" json:"start" binding:"required"`
	End   string `form:"end" json:"end" binding:"required"`
	common.Page
}

type GiftRecordSummary struct {
	Count  int64 `gorm:"column:count"`
	Amount int64 `gorm:"column:amount"`
	Win    int64 `gorm:"column:win"`
}

func (service *GiftSendRequestService) Handle(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")
	user := u.(model.User)

	var gift ploutos.Gift

	if service.StreamerName == "" {
		return serializer.Response{
			Msg: "Invalid Streamer Name",
		}, errors.New("invalid streamer name")
	}

	err = model.DB.Where(`id`, service.GiftId).Find(&gift).Error
	if err != nil || gift.ID == 0 {
		return serializer.Response{
			Msg: "Invalid Gift",
		}, err
	}

	giftRecord := &ploutos.GiftRecord{
		UserId:       user.ID,
		GiftId:       service.GiftId,
		Quantity:     service.Quantity,
		LiveStreamId: service.LiveStreamId,
		TotalPrice:   int64(service.Quantity) * gift.Price,
		StreamerName: service.StreamerName,
	}
	var userSum model.UserSum

	err = model.DB.Transaction(func(tx *gorm.DB) error {

		userSum, _ = model.UserSum{}.GetByUserIDWithLockWithDB(user.ID, tx)
		// userSum.MaxWithdrawable -= giftRecord.TotalPrice

		balanceBefore := userSum.Balance
		balanceAfter := userSum.Balance - giftRecord.TotalPrice

		if balanceAfter < 0 {
			util.GetLoggerEntry(c).Errorf("余额不足")
			return errors.New("余额不足")
		}

		giftRecord.BalanceBefore = balanceBefore
		giftRecord.BalanceAfter = balanceAfter

		giftRecord.TransactionId = strconv.Itoa(int(user.ID)) + strconv.Itoa(int(time.Now().UTC().UnixMilli())) + strconv.Itoa(int(service.GiftId))

		err := tx.Create(&giftRecord).Error
		if err != nil {
			util.GetLoggerEntry(c).Errorf("Send gift failed: %s", err.Error())
			return err
		}

		userSum.Balance -= giftRecord.TotalPrice
		err = tx.Save(&userSum).Error
		if err != nil {
			util.GetLoggerEntry(c).Errorf("User balance not enough: %s", err.Error())
			return err
		}

		transaction := ploutos.Transaction{
			UserId:                user.ID,
			Amount:                giftRecord.TotalPrice,
			BalanceBefore:         balanceBefore,
			BalanceAfter:          balanceAfter,
			TransactionType:       ploutos.TransactionTypeSendGift,
			Wager:                 0,
			WagerBefore:           userSum.RemainingWager,
			WagerAfter:            userSum.RemainingWager,
			ExternalTransactionId: strconv.Itoa(int(giftRecord.ID)),
			GameVendorId:          0,
		}
		err = tx.Create(&transaction).Error
		if err != nil {
			log.Printf("Error creating transaction, err: %v", err)
			return err
		}

		return nil
	})

	if err != nil {
		return serializer.Response{
			Msg: "余额不足",
		}, err
	}

	common.SendGiftSocketMessage(user.ID, gift.ID, service.Quantity, gift.Name, gift.IsAnimated, service.AvatarUrl, user.Nickname, service.LiveStreamId, i18n.T("send_de"), service.VipId, giftRecord.TotalPrice/100)
	common.SendUserSumSocketMsg(user.ID, userSum.UserSum, "send_gift", float64(giftRecord.TotalPrice)/100)
	// common.SendUserSumSocketMsg(3086, userSum.UserSum, "bet", 8819)

	return serializer.Response{
		Msg: i18n.T("success"),
	}, nil
}

func (service *GiftRecordListService) List(c *gin.Context) (r serializer.Response, err error) {
	var giftRecords []ploutos.GiftRecord
	var giftRecordSummary GiftRecordSummary
	i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")
	user := u.(model.User)
	var start, end time.Time
	loc := c.MustGet("_tz").(*time.Location)
	if service.Start != "" {
		if v, e := time.ParseInLocation(time.DateOnly, service.Start, loc); e == nil {
			start = v.UTC()
		}
	}
	if service.End != "" {
		if v, e := time.ParseInLocation(time.DateOnly, service.End, loc); e == nil {
			end = v.UTC().Add(24*time.Hour - 1*time.Second)
		}
	}

	err = model.DB.Model(ploutos.GiftRecord{}).Scopes(model.ByOrderGiftRecordListConditions(user.ID, start, end)).
		Select(`COUNT(1) as count, SUM(total_price) as amount, SUM(total_price) as win`).Find(&giftRecordSummary).Error
	if err != nil {
		return serializer.DBErr(c, service, i18n.T("general_error"), err), err
	}

	err = model.DB.Model(ploutos.GiftRecord{}).Preload("Gift").Scopes(model.ByOrderGiftRecordListConditions(user.ID, start, end), model.ByGiftRecordSort, model.Paginate(service.Page.Page, service.Page.Limit)).
		Find(&giftRecords).Error
	if err != nil {
		return serializer.DBErr(c, service, i18n.T("general_error"), err), err
	}

	r = serializer.Response{
		Data: serializer.BuildPaginatedGiftRecord(giftRecords, giftRecordSummary.Count, giftRecordSummary.Amount, giftRecordSummary.Win),
	}
	return
}
