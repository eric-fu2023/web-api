package common

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"

	"web-api/conf"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/serializer"
	"web-api/util"
	"web-api/websocket"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"firebase.google.com/go/v4/messaging"
	"golang.org/x/text/message"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/plugin/dbresolver"
)

const NOTIFICATION_PLACE_BET_TITLE = "NOTIFICATION_PLACE_BET_TITLE"
const NOTIFICATION_PLACE_BET = "NOTIFICATION_PLACE_BET"
const NOTIFICATION_DEPOSIT_SUCCESS_TITLE = "NOTIFICATION_DEPOSIT_SUCCESS_TITLE"
const NOTIFICATION_DEPOSIT_SUCCESS = "NOTIFICATION_DEPOSIT_SUCCESS"
const NOTIFICATION_WITHDRAWAL_SUCCESS_TITLE = "NOTIFICATION_WITHDRAWAL_SUCCESS_TITLE"
const NOTIFICATION_WITHDRAWAL_SUCCESS = "NOTIFICATION_WITHDRAWAL_SUCCESS"
const NOTIFICATION_WITHDRAWAL_FAILED_TITLE = "NOTIFICATION_WITHDRAWAL_FAILED_TITLE"
const NOTIFICATION_WITHDRAWAL_FAILED = "NOTIFICATION_WITHDRAWAL_FAILED"
const NOTIFICATION_DEPOSIT_BONUS_SUCCESS_TITLE = "NOTIFICATION_DEPOSIT_BONUS_SUCCESS_TITLE"
const NOTIFICATION_DEPOSIT_BONUS_SUCCESS = "NOTIFICATION_DEPOSIT_BONUS_SUCCESS"
const NOTIFICATION_BIRTHDAY_BONUS_SUCCESS_TITLE = "NOTIFICATION_BIRTHDAY_BONUS_SUCCESS_TITLE"
const NOTIFICATION_BIRTHDAY_BONUS_SUCCESS = "NOTIFICATION_BIRTHDAY_BONUS_SUCCESS"
const NOTIFICATION_VIP_PROMOTION_BONUS_TITLE = "NOTIFICATION_VIP_PROMOTION_BONUS_TITLE"
const NOTIFICATION_VIP_PROMOTION_BONUS = "NOTIFICATION_VIP_PROMOTION_BONUS"
const NOTIFICATION_WEEKLY_BONUS_TITLE = "NOTIFICATION_WEEKLY_BONUS_TITLE"
const NOTIFICATION_WEEKLY_BONUS = "NOTIFICATION_WEEKLY_BONUS"
const NOTIFICATION_REBATE_TITLE = "NOTIFICATION_REBATE_TITLE"
const NOTIFICATION_REBATE = "NOTIFICATION_REBATE"
const NOTIFICATION_VIP_PROMOTION_TITLE = "NOTIFICATION_VIP_PROMOTION_TITLE"
const NOTIFICATION_VIP_PROMOTION = "NOTIFICATION_VIP_PROMOTION"
const NOTIFICATION_REFERRAL_ALLIANCE_TITLE = "NOTIFICATION_REFERRAL_ALLIANCE_TITLE"
const NOTIFICATION_REFERRAL_ALLIANCE = "NOTIFICATION_REFERRAL_ALLIANCE"
const NOTIFICATION_POPUP_WINLOSE_FIRST_TITLE = "NOTIFICATION_POPUP_WINLOSE_FIRST_TITLE"
const NOTIFICATION_POPUP_WINLOSE_SECOND_TITLE = "NOTIFICATION_POPUP_WINLOSE_SECOND_TITLE"
const NOTIFICATION_POPUP_WINLOSE_THIRD_TITLE = "NOTIFICATION_POPUP_WINLOSE_THIRD_TITLE"
const NOTIFICATION_POPUP_WIN_FIRST_DESC = "NOTIFICATION_POPUP_WIN_FIRST_DESC"
const NOTIFICATION_POPUP_WIN_SECOND_DESC = "NOTIFICATION_POPUP_WIN_SECOND_DESC"
const NOTIFICATION_POPUP_WIN_THIRD_DESC = "NOTIFICATION_POPUP_WIN_THIRD_DESC"
const NOTIFICATION_POPUP_LOSE_FIRST_DESC = "NOTIFICATION_POPUP_LOSE_FIRST_DESC"
const NOTIFICATION_POPUP_LOSE_SECOND_DESC = "NOTIFICATION_POPUP_LOSE_SECOND_DESC"
const NOTIFICATION_POPUP_LOSE_THIRD_DESC = "NOTIFICATION_POPUP_LOSE_THIRD_DESC"

const NOTIFICATION_SPIN_FIRST_TITLE = "NOTIFICATION_SPIN_FIRST_TITLE"
const NOTIFICATION_SPIN_SECOND_TITLE = "NOTIFICATION_SPIN_SECOND_TITLE"
const NOTIFICATION_SPIN_THIRD_TITLE = "NOTIFICATION_SPIN_THIRD_TITLE"
const NOTIFICATION_SPIN_FIRST_DESC = "NOTIFICATION_SPIN_FIRST_DESC"
const NOTIFICATION_SPIN_SECOND_DESC = "NOTIFICATION_SPIN_SECOND_DESC"
const NOTIFICATION_SPIN_THIRD_DESC = "NOTIFICATION_SPIN_THIRD_DESC"

var (
	ErrInsuffientBalance     = errors.New("insufficient balance or invalid transaction")
	ErrGameVendorUserInvalid = errors.New("user not found in game_vendor_users")
)

type Platform struct {
	Platform int64 `form:"platform" json:"platform" binding:"required"`
}

type Page struct {
	Page  int `form:"page" json:"page" binding:"min=1"`
	Limit int `form:"limit" json:"limit" binding:"min=1"`
}

type PageNoBinding struct {
	Page  int `form:"page" json:"page"`
	Limit int `form:"limit" json:"limit"`
}

type PageById struct {
	IdFrom int `form:"id_from" json:"id_from"`
	Limit  int `form:"limit" json:"limit" binding:"min=1"`
}

type UserRegisterInterface interface {
	CreateUser(model.User, string) error
	VendorRegisterError() error
	OthersError() error
}

// CallbackInterface
type CallbackInterface interface {
	NewCallback(userId int64)
	GetGameVendorId() int64
	GetGameTransactionId() int64
	GetExternalUserId() string
	SaveGameTransaction(*gorm.DB) error
	ShouldProceed() bool
	GetAmount() int64
	GetBetAmountOnly() int64
	GetWagerMultiplier() (factor int64, shouldMultiply bool)
	GetBetAmount() (betPrincipal int64, betExists bool)
	IsAdjustment() bool
	ApplyInsuranceVoucher(userId int64, betAmount int64, betExists bool) error
}

type TestTeamupGamePopUpNotification struct {
	Amount       int64  `json:"amount"`
	Icon         string `json:"icon"`
	ProviderName string `json:"provider_name"`
	EndTime      int64  `json:"end_time"`
	TeamupId     int64  `json:"teamup_id"`
	OrderId      string `json:"order_id"`
	TeamupType   int64  `json:"teamup_type"`
}

func GetUserAndSum(tx *gorm.DB, gameVendor int64, externalUserId string) (gameVendorUser ploutos.GameVendorUser, balance int64, remainingWager int64, maxWithdrawable int64, err error) {
	gameVendorUser, err = GetGameVendorUser(gameVendor, externalUserId)
	if err != nil {
		return
	}
	balance, remainingWager, maxWithdrawable, err = GetSums(tx, gameVendorUser)
	if err != nil {
		return
	}
	return
}

func GetGameVendorUser(vendor int64, userId string) (gpu ploutos.GameVendorUser, err error) {
	err = model.DB.Scopes(model.GameVendorUserByVendorAndExternalUser(vendor, userId)).First(&gpu).Error
	return
}

func GetSums(tx *gorm.DB, gpu ploutos.GameVendorUser) (balance int64, remainingWager int64, maxWithdrawable int64, err error) {
	var userSum ploutos.UserSum
	err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(`user_id`, gpu.UserId).First(&userSum).Error
	if err != nil {
		return
	}
	balance = userSum.Balance
	remainingWager = userSum.RemainingWager
	maxWithdrawable = userSum.MaxWithdrawable
	return
}

func ProcessTransaction(obj CallbackInterface) (err error) {
	tx := model.DB.Clauses(dbresolver.Use("txConn")).Begin(&sql.TxOptions{Isolation: sql.LevelSerializable})
	if tx.Error != nil {
		err = tx.Error
		return
	}

	gpu, balance, remainingWager, maxWithdrawable, err := GetUserAndSum(tx, obj.GetGameVendorId(), obj.GetExternalUserId())
	if err != nil {
		err = fmt.Errorf("%w: %w", ErrGameVendorUserInvalid, err)
		tx.Rollback()
		return
	}
	obj.NewCallback(gpu.UserId)
	if !obj.ShouldProceed() {
		tx.Rollback()
		return
	}
	newBalance := balance + obj.GetAmount()
	if newBalance < 0 {
		err = ErrInsuffientBalance
		tx.Rollback()
		return
	}
	newRemainingWager := remainingWager
	betAmount, betExists, w, e := calWager(obj, remainingWager)
	if e == nil {
		newRemainingWager = w
	}
	newWithdrawable := maxWithdrawable
	if w, e := calMaxWithdrawable(obj, newBalance, newRemainingWager, maxWithdrawable); e == nil {
		newWithdrawable = w
	}
	userSum := ploutos.UserSum{
		Balance:         newBalance,
		RemainingWager:  newRemainingWager,
		MaxWithdrawable: newWithdrawable,
	}
	rows := tx.Select(`balance`, `remaining_wager`, `max_withdrawable`).Where(`user_id`, gpu.UserId).Updates(userSum).RowsAffected
	if rows == 0 {
		err = ErrInsuffientBalance
		tx.Rollback()
		return
	}
	err = obj.SaveGameTransaction(tx)
	if err != nil {
		tx.Rollback()
		return
	}

	// fixme may need to update deposit wager too, since this routine is used by batace in imsb.[OnBalanceDeduction] in batace's environment.
	transaction := ploutos.Transaction{
		UserId:               gpu.UserId,
		Amount:               obj.GetAmount(),
		BalanceBefore:        balance,
		BalanceAfter:         newBalance,
		ForeignTransactionId: obj.GetGameTransactionId(),
		Wager:                userSum.RemainingWager - remainingWager,
		WagerBefore:          remainingWager,
		WagerAfter:           userSum.RemainingWager,
		IsAdjustment:         obj.IsAdjustment(),
		GameVendorId:         obj.GetGameVendorId(),
	}
	err = tx.Save(&transaction).Error
	if err != nil {
		tx.Rollback()
		return
	}
	tx.Commit()

	e = obj.ApplyInsuranceVoucher(gpu.UserId, abs(betAmount), betExists)
	if e != nil {
		util.Log().Error("apply insurance voucher error: ", e.Error())
	}

	//SendNotification(gpu.UserId, consts.Notification_Type_Bet_Placement, NOTIFICATION_PLACE_BET_TITLE, NOTIFICATION_PLACE_BET)
	SendUserSumSocketMsg(gpu.UserId, userSum, "bet", float64(obj.GetAmount())/100)

	return

}

func calWager(obj CallbackInterface, originalWager int64) (betAmount int64, betExists bool, newWager int64, err error) {
	newWager = originalWager

	multiplier, exists := obj.GetWagerMultiplier()
	betAmount, betExists = obj.GetBetAmount()
	if !exists || !betExists {
		return
	}

	absBetAmount := abs(betAmount)
	wager := abs(absBetAmount - abs(obj.GetAmount()))

	if wager > absBetAmount {
		wager = absBetAmount
	}
	newWager = originalWager + (multiplier * wager)

	if newWager < 0 {
		newWager = 0
	}
	return
}

func calMaxWithdrawable(obj CallbackInterface, balance int64, remainingWager int64, originalWithdrawable int64) (newWithdrawable int64, err error) {
	newWithdrawable = originalWithdrawable
	_, exists := obj.GetWagerMultiplier()
	if !exists {
		return
	}
	if remainingWager == 0 {
		//if balance > originalWithdrawable {
		newWithdrawable = balance
		//}
	}
	return
}

func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

func SendNotification(userId int64, notificationType string, title string, text string, v ...serializer.Response) { // add to notification list and push
	go func() {
		notification := ploutos.UserNotification{
			UserId: userId,
			Text:   text,
		}
		if err := notification.Send(model.DB); err != nil {
			util.Log().Error("notification creation error: ", err.Error())
			return
		}
	}()

	// check whether the v argument is provided with a value
	if len(v) > 0 {
		PushNotification(userId, notificationType, title, text, v[0])
	} else {
		PushNotification(userId, notificationType, title, text)
	}
}

func PushNotification(userId int64, notificationType string, title string, text string, v ...serializer.Response) {
	go func() {
		var msgData map[string]string
		if len(v) > 0 {
			resp, _ := json.Marshal(v[0].Data)
			msgData = map[string]string{
				"notification_type": notificationType,
				"data":              string(resp),
			}
		} else {
			msgData = map[string]string{
				"notification_type": notificationType,
			}
		}

		notification := messaging.Notification{
			Title: title,
			Body:  text,
		}
		tokens, err := model.GetFcmTokenStrings([]int64{userId})
		if err != nil {
			util.Log().Error("fcm token generation error: ", err.Error())
			return
		}
		client := util.FCMFactory(false)
		err = client.SendMessageToAll(context.TODO(), msgData, notification, tokens)
		if err != nil {
			util.Log().Error("fcm sending error: ", err.Error())
			return
		}
	}()
}

func PushNotificationAll(notificationType string, title string, text string, v ...serializer.Response) {
	go func() {
		var msgData map[string]string
		if len(v) > 0 {
			resp, _ := json.Marshal(v[0].Data)
			msgData = map[string]string{
				"notification_type": notificationType,
				"data":              string(resp),
			}
		} else {
			msgData = map[string]string{
				"notification_type": notificationType,
			}
		}

		notification := messaging.Notification{
			Title: title,
			Body:  text,
		}
		var user_ids []int64
		model.DB.Table("users").Select("id").Scan(&user_ids)
		tokens, err := model.GetFcmTokenStrings(user_ids)
		if err != nil {
			util.Log().Error("fcm token generation error: ", err.Error())
			return
		}
		client := util.FCMFactory(false)
		err = client.SendMessageToAll(context.TODO(), msgData, notification, tokens)
		if err != nil {
			util.Log().Error("fcm sending error: ", err.Error())
			return
		}
	}()
}

func LogGameCallbackRequest(action string, request any) {
	j, _ := json.Marshal(request)
	util.Log().Info(`%s: %s`, action, string(j))
}

func SendCashNotification(userId int64, notificationType string, title string, text string, amount int64, currencyId int64) {
	go func() {
		var currency ploutos.Currency
		err := model.DB.Where(`id`, currencyId).First(&currency).Error
		if err != nil {
			util.Log().Error("send cash notification error (query currency): ", err.Error())
			return
		}

		lang, err := model.GetUserLang(userId)
		if err != nil {
			util.Log().Error("send cash notification error (GetUserLang fail): ", err.Error())
			return
		}
		title = conf.GetI18N(lang).T(title)
		text = conf.GetI18N(lang).T(text)
		p := message.NewPrinter(message.MatchLanguage(lang))
		SendNotification(userId, notificationType, title, p.Sprintf(text, float64(amount)/100, currency.Name))
	}()
}

func SendCashNotificationWithoutCurrencyId(userId int64, notificationType string, title string, text string, amount int64) {
	go func() {
		var user ploutos.User
		err := model.DB.Where(`id`, userId).First(&user).Error
		if err != nil {
			util.Log().Error("send cash notification error (query user): ", err.Error())
			return
		}
		SendCashNotification(userId, notificationType, title, text, amount, user.CurrencyId)
	}()
}

func SendUserSumSocketMsg(userId int64, userSum ploutos.UserSum, cause string, amount float64) {
	go func() {
		conn := websocket.Connection{}
		conn.Connect(os.Getenv("WS_NOTIFICATION_URL"), os.Getenv("WS_NOTIFICATION_TOKEN"), []func(*websocket.Connection, context.Context, context.CancelFunc){
			func(conn *websocket.Connection, ctx context.Context, cancelFunc context.CancelFunc) {
				select {
				case <-ctx.Done():
					return
				default:
					cause_type := cause
					if cause_type == "FTD_success" {
						cause_type = "deposit_success"
					}
					msg := websocket.BalanceUpdateMessage{
						Room:  serializer.UserSignature(userId),
						Event: "balance_change",
						Cause: cause_type,
						// Cause:           "deposit_success", // this is requested by FE
						Amount:          amount,
						Balance:         float64(userSum.Balance) / 100,
						RemainingWager:  float64(userSum.RemainingWager) / 100,
						MaxWithdrawable: float64(userSum.MaxWithdrawable) / 100,
						IsFTD:           cause == "FTD_success",
					}
					msg.Send(conn)
					util.Log().Info("deposit callback msg sent for %d, cause_type: %s, IsFTD: %b, Amount: %2f ", userId, cause_type, cause == "FTD_success", amount)
				}
			},
		})
	}()
}

func SendGiftSocketMessage(userId int64, giftId int64, giftQuantity int, giftName string, isGiftAnimated bool, avatar string, nickname string, liveStreamId int64, message string, vipId int64, totalPrice int64) {
	go func() {
		conn := websocket.Connection{}
		conn.Connect(os.Getenv("WS_URL"), os.Getenv("WS_TOKEN"), []func(*websocket.Connection, context.Context, context.CancelFunc){
			func(conn *websocket.Connection, ctx context.Context, cancelFunc context.CancelFunc) {
				select {
				case <-ctx.Done():
					return
				default:
					msg := websocket.RoomMessage{
						Room:           fmt.Sprintf(`stream:%d`, liveStreamId),
						Message:        message + " " + giftName + " " + strconv.Itoa(giftQuantity) + " x ",
						UserId:         userId,
						UserType:       consts.ChatUserType["user"],
						Nickname:       nickname,
						Avatar:         avatar,
						Type:           consts.WebSocketMessageType["gift"],
						GiftId:         giftId,
						GiftQuantity:   giftQuantity,
						GiftName:       giftName,
						IsAnimated:     isGiftAnimated,
						VipId:          vipId,
						TotalGiftPrice: totalPrice,
					}
					if e := msg.Send(conn); e != nil {
						cancelFunc()
						return
					}
				}
			},
		})
	}()
}

func SendTeamupGamePopupNotificationSocketMsg(userId int64, teamupId, endTime, amount, teamupType int64, providerName, icon, orderId string) {
	go func() {
		conn := websocket.Connection{}
		conn.Connect(os.Getenv("WS_NOTIFICATION_URL"), os.Getenv("WS_NOTIFICATION_TOKEN"), []func(*websocket.Connection, context.Context, context.CancelFunc){
			func(conn *websocket.Connection, ctx context.Context, cancelFunc context.CancelFunc) {
				select {
				case <-ctx.Done():
					return
				default:
					msg := websocket.TeamupGameNotificationMessage{
						Room:         serializer.UserSignature(userId),
						UserId:       userId,
						Event:        "teamup_game",
						TeamupId:     teamupId,
						EndTime:      endTime,
						Amount:       amount,
						ProviderName: providerName,
						Icon:         icon,
						TeamupType:   teamupType,
						OrderId:      orderId,
					}
					msg.Send(conn)
				}
			},
		})
	}()
}

func TESTSENDCHAT(userId int64) {
	go func() {
		conn := websocket.Connection{}
		conn.Connect(os.Getenv("WS_URL"), os.Getenv("WS_TOKEN"), []func(*websocket.Connection, context.Context, context.CancelFunc){
			func(conn *websocket.Connection, ctx context.Context, cancelFunc context.CancelFunc) {
				select {
				case <-ctx.Done():
					return
				default:
					msg := websocket.RoomMessage{
						Room:     fmt.Sprintf(`stream:%d`, 3345),
						UserId:   userId,
						UserType: consts.ChatUserType["user"],
					}
					if e := msg.Send(conn); e != nil {
						cancelFunc()
						return
					}
				}
			},
		})
	}()
}

func TESTSENDNOTIFICATION(userId int64) {
	go func() {
		conn := websocket.Connection{}
		conn.Connect(os.Getenv("WS_NOTIFICATION_URL"), os.Getenv("WS_NOTIFICATION_TOKEN"), []func(*websocket.Connection, context.Context, context.CancelFunc){
			func(conn *websocket.Connection, ctx context.Context, cancelFunc context.CancelFunc) {
				select {
				case <-ctx.Done():
					return
				default:
					msg := websocket.TeamupGameNotificationMessage{
						Room:         serializer.UserSignature(userId),
						UserId:       userId,
						Event:        "teamup_game",
						TeamupId:     1,
						EndTime:      1727530027,
						Amount:       68,
						ProviderName: "ADAS",
						Icon:         "",
						OrderId:      "50001829001",
					}
					msg.Send(conn)
				}
			},
		})
	}()
}

func TESTSENDPUSHNOTIFICATION(userId int64) {
	var resp serializer.Response
	resp.Data = TestTeamupGamePopUpNotification{
		TeamupId:     1,
		EndTime:      1727530027,
		Amount:       68,
		ProviderName: "ADAS",
		Icon:         "",
		OrderId:      "50001829001",
		TeamupType:   3,
	}
	SendNotification(userId, "popup_teamup_game", "HEHEHEH", "HAHAHHAHAHHA", resp)
}
