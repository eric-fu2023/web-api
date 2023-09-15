package model

import models "blgit.rfdev.tech/taya/ploutos-object"

type CashOutRule struct{
	models.CashOutRuleC
}

func (CashOutRule) Get(vipLevel int64) (rule CashOutRule, err error) {
	err = DB.Order("vip_level desc").First(&rule).Error
	return
}