package promotion

import (
	"blgit.rfdev.tech/taya/common-function/domain/transactions"
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
)

var promotionTypeToCashOrderType = map[int64]int64{
	ploutos.PromotionTypeFirstDepB:                  ploutos.CashOrderTypeDepB,
	ploutos.PromotionTypeReDepB:                     ploutos.CashOrderTypeDepB,
	ploutos.PromotionTypeFirstDepIns:                ploutos.CashOrderTypeBetIns,
	ploutos.PromotionTypeReDepIns:                   ploutos.CashOrderTypeBetIns,
	ploutos.PromotionTypeBeginnerB:                  ploutos.CashOrderTypeBeginnerB,
	ploutos.PromotionTypeOneTimeDepB:                ploutos.CashOrderTypeDepB,
	ploutos.PromotionTypeVipRebate:                  ploutos.CashOrderTypeVipRebate,
	ploutos.PromotionTypeVipBirthdayB:               ploutos.CashOrderTypeVipBday,
	ploutos.PromotionTypeVipPromotionB:              ploutos.CashOrderTypeVipPromo,
	ploutos.PromotionTypeVipWeeklyB:                 ploutos.CashOrderTypeVipWeekly,
	ploutos.PromotionTypeVipReferral:                ploutos.CashOrderTypeVipReferral,
	ploutos.PromotionTypeCustomTemplate:             ploutos.CashOrderTypeCustomPromotion,
	ploutos.PromotionTypeTeamup:                     ploutos.CashOrderTypeTeamupPromotion,
	ploutos.PromotionTypeCashMethodDepositPromotion: ploutos.CashOrderTypeCashMethodPromotion,
}

var promotionTypeToTransactionTypeMapping = transactions.PromotionTypeToTransactionTypeMapping

const (
	MatchTypeEnded        = 0
	MatchTypePostponed    = 1
	MatchTypeInterrupted  = 2
	MatchTypeCancelled    = 3
	MatchTypeNotStarted   = 4
	MatchTypeLive         = 5
	MatchTypeDelayed      = 6
	MatchTypeAbandoned    = 7
	MatchTypeSuspended    = 8
	MatchTypeCoverageLost = 9

	OddsFormatEU = 1
	OddsFormatHK = 2
)

var imMatchTypeMapping = map[int]int{
	1: MatchTypeNotStarted,
	2: MatchTypeNotStarted,
}
