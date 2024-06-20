package task

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"web-api/model"

	"blgit.rfdev.tech/taya/payment-service/finpay"
	models "blgit.rfdev.tech/taya/ploutos-object"
)

const specialUpdatedBy = -10086

func RefreshPaymentOrder() {
	var orders []model.CashOrder
	err := model.DB.
		Where("status", models.CashOrderStatusPending).
		Where("order_type", 1).
		Where("NOT is_manual_operation").
		Where("created_at < ?", time.Now().Add(-30*time.Minute)).
		Find(&orders).Error
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("polling payment orders")
	cl := finpay.FinpayClient{}
	for _, t := range orders {
		data, err := cl.DefaultPaymentQuery(context.Background(), t.ID)
		serializedData, _ := json.Marshal(data)
		if err != nil {
			fmt.Println(err)
			if errors.Is(err, finpay.ErrorOrderNotFound) {
				err = model.DB.Model(&t).Updates(map[string]any{
					"status":           models.CashOrderStatusFailed,
					"notes":            models.EncryptedStr(serializedData),
					"manual_closed_by": specialUpdatedBy,
				}).Error

				if err == nil {
					fmt.Println("failed order", t.ID)
				} else {
					fmt.Println("failed order err", err)
				}
			}
			continue
		}

		if data.PaymentOrderStatus == "COMPLETED" {
			fmt.Println("completed order, ", t.ID)
			// _, err = cashout.ManualCloseCashOutTxn(context.Background(), t.ID, specialUpdatedBy, string(serializedData))
			// if err == nil {
			// 	fmt.Println("completed order", t.OrderNumber)
			// } else {
			// 	fmt.Println("completed order err", err)
			// }
		} else if data.PaymentOrderStatus == "FAILED" {
			err = model.DB.Debug().Model(&t).Updates(map[string]any{
				"status":           models.CashOrderStatusFailed,
				"notes":            models.EncryptedStr(serializedData),
				"manual_closed_by": specialUpdatedBy,
			}).Error
			if err == nil {
				fmt.Println("failed order", t.ID)
			} else {
				fmt.Println("failed order err", err)
			}
		} else if data.PaymentOrderStatus == "CLOSED" {
			err = model.DB.Debug().Model(&t).Updates(map[string]any{
				"status":           models.CashOrderStatusExpired,
				"notes":            models.EncryptedStr(serializedData),
				"manual_closed_by": specialUpdatedBy,
			}).Error
			if err == nil {
				fmt.Println("expired order", t.ID)
			} else {
				fmt.Println("expired order err", err)
			}
		}
	}
}
