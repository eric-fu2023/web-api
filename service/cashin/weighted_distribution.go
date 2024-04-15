package cashin

import models "blgit.rfdev.tech/taya/ploutos-object"

func GetNextChannel(list []models.CashMethodChannel) models.CashMethodChannel {
	distribution := map[int64]int64{}
	var weightTotal int64 = 0
	accumulation := map[int64]int64{}
	var calledTotal int64 = 0
	for _, item := range list {
		distribution[item.ID] = item.Weight
		weightTotal += item.Weight
		accumulation[item.ID] = item.Stats.Called
		calledTotal += item.Stats.Called
	}
	for _, item := range list {
		if calledTotal == 0 {
			return item
		}
		if float64(accumulation[item.ID])/float64(calledTotal) < float64(distribution[item.ID])/float64(weightTotal) {
			return item
		}
	}
	return list[0]
}
