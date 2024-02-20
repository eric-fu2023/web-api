package promotion

import models "blgit.rfdev.tech/taya/ploutos-object"

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

var promotionOrderTypeMapping = map[int64]int64{
	models.PromotionTypeFirstDepB:   models.CashOrderTypeDepB,
	models.PromotionTypeReDepB:      models.CashOrderTypeDepB,
	models.PromotionTypeFirstDepIns: models.CashOrderTypeBetIns,
	models.PromotionTypeReDepIns:    models.CashOrderTypeBetIns,
	models.PromotionTypeBeginnerB:   models.CashOrderTypeBeginnerB,
}

var promotionTxTypeMapping = map[int64]int64{
	models.PromotionTypeFirstDepB:   models.TransactionTypeDepB,
	models.PromotionTypeReDepB:      models.TransactionTypeDepB,
	models.PromotionTypeFirstDepIns: models.TransactionTypeBetIns,
	models.PromotionTypeReDepIns:    models.TransactionTypeBetIns,
	models.PromotionTypeBeginnerB:   models.TransactionTypeBeginnerB,
	models.PromotionTypeOneTimeDepB: models.TransactionTypeDepB,
}

var imMatchTypeMapping = map[int]int{
	1: MatchTypeNotStarted,
	2: MatchTypeNotStarted,
}
