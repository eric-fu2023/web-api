package on_cash_orders

import (
	"blgit.rfdev.tech/taya/common-function/rfcontext"
	"context"
	"fmt"
	"log"
	"web-api/model"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
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
	PaymentGatewayDefault PaymentGateway = "default"
	PaymentGatewayFinpay  PaymentGateway = "finpay"
	PaymentGatewayForay   PaymentGateway = "foray"
)

type RequestIngressMode = int64

const (
	RequestModeCallback RequestIngressMode = 1
	// RequestModeManual via internal routes (?)
	RequestModeManual RequestIngressMode = 2
)

// Handle
func Handle(ctx context.Context, order model.CashOrder, transactionType ploutos.TransactionType, eventType CashOrderEventType, gateway PaymentGateway, requestMode RequestIngressMode) error {
	ctx = rfcontext.AppendCallDesc(ctx, "on_cash_order.Handle()")
	ctx = rfcontext.AppendParams(ctx, "on_cash_order.Handle()", map[string]interface{}{
		"transactionType": transactionType,
		"eventType":       eventType,
		"gateway":         gateway,
		"requestMode":     requestMode,
		"order":           order,
	})

	// validate eventType
	switch eventType {
	case CashOrderEventTypeClose:
	default:
		return fmt.Errorf("unsupported event type: %d", eventType)
	}
	// validate payment channel
	switch gateway {
	case PaymentGatewayFinpay, PaymentGatewayForay:
	default:
		return fmt.Errorf("unsupported gateway: %s", gateway)
	}
	// one time bonus if it's cash in
	if transactionType == ploutos.TransactionTypeCashIn {
		OneTimeBonusPromotion(ctx, order)
	}
	// handle cash method promotion
	{
		shouldHandleCashMethodPromotion := false
		switch {
		case ploutos.TransactionTypeCashIn == transactionType:
			shouldHandleCashMethodPromotion = true
		case ploutos.TransactionTypeCashOut == transactionType && RequestModeCallback == requestMode:
			shouldHandleCashMethodPromotion = true
		default:
			return fmt.Errorf("unknown transaction type for shouldHandleCashMethodPromotion %d", transactionType)
		}

		ctx = rfcontext.AppendParams(ctx, "cash_method_promo", map[string]interface{}{
			"shouldHandleCashMethodPromotion": shouldHandleCashMethodPromotion,
		})

		log.Printf(rfcontext.Fmt(ctx))

		if shouldHandleCashMethodPromotion {
			ValidateAndClaimCashMethodPromotion(ctx, order)
		}
	}

	return nil
}
