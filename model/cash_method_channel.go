package model

import (
	"context"
	"errors"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
)

func GetNextCashMethodChannel(list []ploutos.CashMethodChannel) (ploutos.CashMethodChannel, error) {
	if len(list) == 0 {
		return ploutos.CashMethodChannel{}, errors.New("no cash_method_channel to filter next channel")
	}
	if len(list) == 1 {
		return list[0], nil
	}
	weight := map[int64]int64{}
	var weightTotal int64 = 0
	called := map[int64]int64{}
	var calledTotal int64 = 0
	for _, item := range list {
		weight[item.ID] = item.Weight
		weightTotal += item.Weight
		called[item.ID] = item.Stats.Called
		calledTotal += item.Stats.Called
	}
	for _, item := range list {
		if calledTotal == 0 || weightTotal == 0 {
			return item, nil
		}
		if float64(called[item.ID])/float64(calledTotal) < float64(weight[item.ID])/float64(weightTotal) {
			return item, nil
		}
	}
	return list[0], nil
}

func FilterCashMethodChannelsByVip(c context.Context, user User, chns []ploutos.CashMethodChannel) []ploutos.CashMethodChannel {
	ret := []ploutos.CashMethodChannel{}
	vip, _ := GetVipWithDefault(c, user.ID)
	for _, ch := range chns {
		for _, lvl := range ch.VipLevels {
			if vip.VipRule.VIPLevel == int64(lvl) {
				ret = append(ret, ch)
				break
			}
		}
	}
	return ret
}

func FilterCashMethodChannelsByAmount(c context.Context, amount int64, chns []ploutos.CashMethodChannel) []ploutos.CashMethodChannel {
	ret := []ploutos.CashMethodChannel{}
	for _, ch := range chns {
		if amount <= ch.MaxAmount && amount >= ch.MinAmount {
			ret = append(ret, ch)
		}
	}
	return ret
}
