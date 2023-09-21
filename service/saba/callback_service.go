package saba

import (
	models "blgit.rfdev.tech/taya/ploutos-object"
	"gorm.io/gorm"
	"web-api/conf/consts"
	"web-api/model"
)

type Callback struct {
	Transaction models.SabaTransactionC
}

func (c *Callback) GetGameProviderId() int64 {
	return consts.GameProvider["saba"]
}

func (c *Callback) GetGameTransactionId() int64 {
	return c.Transaction.ID
}

func (c *Callback) SaveGameTransaction(tx *gorm.DB) error {
	return tx.Save(&c.Transaction).Error
}

func (c *Callback) ShouldProceed() bool {
	return true // saba placebet doesn't have bet that shouldn't proceed
}

func (c *Callback) GetAmount() int64 {
	multiplier := int64(1)
	if c.Transaction.DebitAmount > 0 {
		multiplier = -1
	}
	return multiplier * c.Transaction.ActualAmount
}

func (c *Callback) GetWagerMultiplier() (int64, bool) {
	return -1, true
}

func (c *Callback) GetBetAmount() (amount int64, exists bool) {
	e := model.DB.Model(models.SabaTransactionC{}).Select(`actual_amount`).
		Where(`ref_id`, c.Transaction.RefId).Order(`id`).First(&amount).Error
	if e == nil {
		exists = true
	}
	return
}

func (c *Callback) IsAdjustment() bool {
	return false
}
