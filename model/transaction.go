package model

import (
	"time"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
)

type Transaction struct {
	ploutos.Transaction
}

func (Transaction) ListTxRecord(c *gin.Context, userID int64, timeFrom, timeTo *time.Time) (r []Transaction, err error) {
	q := DB.Where("user_id", userID)
	if timeFrom != nil {
		q = q.Where("created_at > ?", *timeFrom)
	}
	if timeTo != nil {
		q = q.Where("created_at < ?", *timeTo)
	}
	err = q.Find(&r).Error
	return
}
