package model

import ploutos "blgit.rfdev.tech/taya/ploutos-object"

type CashOutRule struct {
	ploutos.CashOutRule
}

func (CashOutRule) Get(vipLevel int64) (rule CashOutRule, err error) {
	err = DB.Order("vip_level desc").Where("vip_level <= ?", vipLevel).Where("is_active").First(&rule).Error
	return
}
