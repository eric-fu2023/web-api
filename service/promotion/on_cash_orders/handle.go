package on_cash_orders

import (
	"context"
	"fmt"

	"web-api/model"

	models "blgit.rfdev.tech/taya/ploutos-object"
)

// Note: Work in progress
// CashOrderEventType, PaymentGateway is a workaround to conditionally trigger promotions. See [Handle]
// may be possible to derive args from cash order.

// CashOrderEventType
type CashOrderEventType = int64

const (
	CashOrderEventTypeClose CashOrderEventType = 2
)

type PaymentGateway = string

const (
	PaymentGatewayFinPay PaymentGateway = "finpay"
	PaymentGatewayForay  PaymentGateway = "foray"
)

type RequestIngressMode = int64

const (
	RequestModeCallback RequestIngressMode = 1
	// RequestModeManual via internal routes (?)
	RequestModeManual RequestIngressMode = 2
)

// Handle
// Note: take awareness on the trigger conditions and also the sequence of process~ these will change as requirement comes in
// See also: [Note: Work in progress]
func Handle(ctx context.Context, order model.CashOrder, transactionType models.TransactionType, eventType CashOrderEventType, gateway PaymentGateway, requestMode RequestIngressMode) error {
	switch eventType {
	case CashOrderEventTypeClose:
	default:
		return fmt.Errorf("unsupported event type %d", eventType)
	}

	switch gateway {
	case PaymentGatewayFinPay, PaymentGatewayForay:
	default:
		return fmt.Errorf("unsupported gateway %s", gateway)
	}

	{
		var shouldHandleOneTimeBonus bool
		shouldHandleOneTimeBonus = transactionType == models.TransactionTypeCashIn

		if shouldHandleOneTimeBonus {
			OneTimeBonusPromotion(ctx, order)
		}
	}

	{
		shouldHandleCashMethodPromotion := false
		switch {
		case models.TransactionTypeCashIn == transactionType:
			shouldHandleCashMethodPromotion = true
		case models.TransactionTypeCashOut == transactionType && RequestModeCallback == requestMode:
			shouldHandleCashMethodPromotion = true
		default:
			return fmt.Errorf("unknown transaction type for shouldHandleCashMethodPromotion %d", transactionType)
		}

		if shouldHandleCashMethodPromotion {
			CashMethodPromotion(ctx, order)
		}
	}

	{
		shouldHandleTetheredRebatePromotion := false
		_ = shouldHandleTetheredRebatePromotion
		// trpService, err := tethered_rebate_promotion.NewService(model.DB, nil, nil)
		// if err != nil {
		// 	util.Log().Error("tethered_rebate_promotion.NewService failed", err)
		// }
		// cashOrder := order.CashOrder
		// _ = cashOrder
		// reward, err := trpService.AddRewardForClosedDeposit(context.TODO(), tethered_rebate_promotion.UserForm{
		// 	Id: 0,
		// }, nil)

		// _ = reward

	}
	return nil
}
