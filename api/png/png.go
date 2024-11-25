package png

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"

	"web-api/api"
	"web-api/model"

	"blgit.rfdev.tech/taya/common-function/rfcontext"
	"blgit.rfdev.tech/taya/game-service/png/callback"
	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func CasinoGamesSessionOpen(ctx context.Context, message MessageT) (callback.Message_CasinoGamesSessionOpen, error) {
	return callback.Message_CasinoGamesSessionOpen(message), nil
}

// MessageT is all-fields-encompassing and MessageType-blind.
// [ ] CasinoTransactionReleaseOpen
// [ ] CasinoPlayerLogin, CasinoPlayerLogout ... // TODO
type MessageT struct {
	TransactionId         int64                      `json:"TransactionId"`
	Status                callback.IntegrationStatus `json:"Status"`
	Amount                float64                    `json:"Amount"`
	Time                  callback.ISO8601           `json:"Time"`
	ProductGroup          int64                      `json:"ProductGroup"`
	ExternalUserId        string                     `json:"ExternalUserId"`
	GameSessionId         int64                      `json:"GamesessionId"`
	GameSessionState      callback.GameSessionState  `json:"GamesessionState"`
	GameId                int64                      `json:"GameId"`
	RoundId               int64                      `json:"RoundId"`
	RoundData             interface{}                `json:"RoundData"`
	RoundLoss             float64                    `json:"RoundLoss"`
	JackpotLoss           float64                    `json:"JackpotLoss"`
	JackpotGain           float64                    `json:"JackpotGain"`
	Currency              string                     `json:"Currency"`
	ExternalTransactionId string                     `json:"ExternalTransactionId"`
	Balance               float64                    `json:"Balance"`
	NumRounds             int64                      `json:"NumRounds"`
	TotalLoss             float64                    `json:"TotalLoss"`
	TotalGain             float64                    `json:"TotalGain"`
	ExternalFreeGameId    interface{}                `json:"ExternalFreegameId"`
	Channel               callback.Channel           `json:"Channel"`
	MessageId             *string                    `json:"MessageId"`
	MessageType           callback.MessageType       `json:"MessageType"`
	MessageTimestamp      string                     `json:"MessageTimestamp"`
}

type Request struct {
	Messages []MessageT `json:"Messages"`
}

// Consume
// TODO recommended to use queue instead.
func Consume(ctx context.Context, req Request) error {
	ctx = rfcontext.AppendCallDesc(ctx, "Consume")
	messages := req.Messages

	// any err => return non-nil
	var scopeErr error

	messages_SessionOpen := []callback.Message_CasinoGamesSessionOpen{}
	for _, message := range messages {
		switch message.MessageType {
		case callback.MessageTypeCasinoGamesSessionOpen:
			mT, errT := CasinoGamesSessionOpen(ctx, message)
			if errT != nil {
				scopeErr = errors.Join(scopeErr, errT)
				log.Println(rfcontext.FmtJSON(rfcontext.AppendError(ctx, errT, "CasinoGamesSessionOpen()")))
			}
			messages_SessionOpen = append(messages_SessionOpen, mT)
		}
	}

	{ // MessageTypeCasinoGamesSessionOpen
		reports, oErr := OnMessageCasinoGamesSessionOpen(ctx, messages_SessionOpen)
		if oErr != nil {
			scopeErr = errors.Join(scopeErr, oErr)
			log.Println(rfcontext.FmtJSON(rfcontext.AppendError(ctx, oErr, "OnMessageCasinoGamesSessionOpen()")))
		}
		err := InsertReports(reports)
		if err != nil {
			log.Println(rfcontext.FmtJSON(rfcontext.AppendError(ctx, err, "InsertReports()")))
			scopeErr = errors.Join(scopeErr, err)
		}
	}
	return scopeErr
}

func IntegrationStatusToReportStatus(status callback.IntegrationStatus) (ploutos.TayaBetReportStatus, error) {
	switch status {
	case callback.IntegrationStatusSuccessful:
		return ploutos.TayaBetReportStatusSettled, nil
	case callback.IntegrationStatusFailed:
		return ploutos.TayaBetReportStatusCreated, nil

	case callback.IntegrationStatusVoided:
		return ploutos.TayaBetReportStatusCancelled, nil
	}
	return 0, errors.New("unknown mapping for IntegrationStatusToReportStatus")
}

func ToReport(message callback.Message_CasinoGamesSessionOpen) (ploutos.PNGBetReport, error) {
	roundId := strconv.FormatInt(message.RoundId, 10)

	userId, err := strconv.Atoi(message.ExternalUserId)
	if err != nil {
		return ploutos.PNGBetReport{}, err
	}
	_totalLoss := message.TotalLoss
	totalLoss := int64(_totalLoss * 100)
	turnover := int64(_totalLoss * 100)

	pl := int64(message.Amount * 100)
	totalGain := int64(message.TotalGain * 100)

	status, sErr := IntegrationStatusToReportStatus(message.Status)
	if sErr != nil {
		return ploutos.PNGBetReport{}, sErr
	}

	betTime, tErr := message.Time.ToTime()
	if tErr != nil {
		return ploutos.PNGBetReport{}, tErr
	}

	rawLog, err := json.Marshal(message)
	if err != nil {
		fmt.Println("error raw encoding betLog:", err)
		return ploutos.PNGBetReport{}, err
	}

	newUUID := uuid.NewString()
	return ploutos.PNGBetReport{
		BASE_UUID: ploutos.BASE_UUID{
			ID: &newUUID,
		},
		OrderId:      "PNG" + roundId,
		BusinessId:   roundId,
		UserId:       int64(userId),
		GameType:     message.GameId,
		Bet:          totalLoss,
		Wager:        turnover,
		Win:          totalGain, // payout
		ProfitLoss:   pl,
		Status:       status,
		BetTime:      &betTime,
		RewardTime:   &betTime,
		InfoJson:     rawLog,
		IsParlay:     false,
		BetType:      "",
		MatchCount:   message.NumRounds,
		MaxWinAmount: "",
		GameId:       ploutos.GAME_INTEGRATION_PNG,
		Provider:     "",
		RefId:        0,
		WagerSettled: false,
	}, nil
}
func OnMessageCasinoGamesSessionOpen(ctx context.Context, messages []callback.Message_CasinoGamesSessionOpen) ([]ploutos.PNGBetReport, error) {
	ctx = rfcontext.AppendCallDesc(ctx, "OnMessageCasinoGamesSessionOpen")

	reports := []ploutos.PNGBetReport{}
	for _, m := range messages {
		report, err := ToReport(m)
		if err != nil {
			log.Println(rfcontext.FmtJSON(rfcontext.AppendError(ctx, err, "ToReport()")))
		}

		reports = append(reports, report)
	}

	return reports, nil
}

// BetReportUniqueColumns copied from task system
var BetReportUniqueColumns = []clause.Column{{Name: "business_id"}, {Name: "user_id"}, {Name: "game_type"}}

func InsertReports(_reportsToCreate []ploutos.PNGBetReport) error {
	if len(_reportsToCreate) == 0 {
		return nil
	}
	err := model.DB.Session(&gorm.Session{CreateBatchSize: 200}).Clauses(
		clause.OnConflict{
			Columns:   BetReportUniqueColumns,
			UpdateAll: true,
			Where: clause.Where{Exprs: []clause.Expression{
				clause.Neq{Column: "png_bet_reports.status", Value: ploutos.TayaBetReportStatusSettled},
			}},
		},
	).Create(&_reportsToCreate).Error
	return err
}

// Feed
// single controller endpoint for all push messages
func Feed(c *gin.Context) {
	var req Request
	if bErr := c.ShouldBind(&req); bErr == nil {
		ctx := rfcontext.AppendCallDesc(context.Background(), "png.Feed")
		if err := Consume(ctx, req); err != nil {
			c.JSON(500, api.ErrorResponse(c, req, err))
		}
	} else {
		c.JSON(400, api.ErrorResponse(c, req, bErr))
	}
}
