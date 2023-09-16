package cashout

import "web-api/model"

func CalTxDetails(list []model.Transaction) (totalOut, payoutCount int64) {
	for _, item := range list {
		if item.TransactionType == 10001 {
			totalOut += item.Amount
			payoutCount += 1
		}
	}
	totalOut = -totalOut
	return
}
