// Package tethered_rebate_promotion
// promotion: referrer to get rebate as reward on referree's first 3 deposits
package tethered_rebate_promotion

import (
	"context"
	"errors"
	"time"

	"web-api/util"

	po "blgit.rfdev.tech/taya/ploutos-object"

	"gorm.io/gorm"
)

type ReadLimitFunc func() int64

// Service
// To be called after relevant records (cash order, transaction, cash method etc) are updated in the database
type Service struct {
	db                         *gorm.DB
	depositRankCountReadLimitF ReadLimitFunc
	referrerRebateF            func(orderType po.CashOrderOrderType, operationType po.OperationType, amount float64)
}

func NewService(db *gorm.DB,
	depositRankCountReadLimitF ReadLimitFunc,
	referrerRebateF func(orderType po.CashOrderOrderType, operationType po.OperationType)) (*Service, error) {
	if db == nil {
		return nil, errors.New("no gorm")
	}
	if depositRankCountReadLimitF == nil {
		return nil, errors.New("no depositRankCountReadLimitF")
	}
	if referrerRebateF == nil {
		return nil, errors.New("no referrerRebateF")
	}
	return &Service{}, nil
}

// AddRewardForClosedDeposit
// adds redeemable reward for the referrer
//
// Example User Flow
// 1. A refers B
// 2. B deposits (1st, 2nd....)
// 3. Server add reward for A
func (s *Service) AddRewardForClosedDeposit(ctx context.Context, referee UserForm, refereeCashOrder *RefereeCashOrder) (any, error) {
	panic("implement me")
	if refereeCashOrder == nil {
		return nil, errors.New("referee cash order is nil")
	}
	refereeId := referee.Id
	// 1. prevalidate if need to add reward.
	// 1.1 Get referee's deposit count ranked by updated (order type 1, operation type 0 or 3000 to 3999)

	// - read back limit by rank count
	_ = s.DepositRankCountReadLimit()

	db := s.db
	var depositCashOrders []RefereeCashOrder
	// select deposit cashorders by userid, limit  by dense rank
	_ = db.Debug().Table("cash_order").Find(&depositCashOrders).Error

	// 1.2
	// find if there's a match

	var selectedCashOrder *RefereeCashOrder
	for _, c := range depositCashOrders {
		if c.ID == refereeCashOrder.ID {
			selectedCashOrder = &c
			break
		}
	}

	if selectedCashOrder == nil {
		util.Log().Debug("no reward applicable. cash order not in depositCashOrders")
		return nil, nil
	}

	// 2. match found. check if referred.

	var referrer Referrer

	rErr := db.Debug().Table("user_referrals").
		Joins("LEFT JOIN users ON users.id = user_referrals.referral_id").
		Where("user_referrals.referral_id = ?", refereeId).
		Select("users.id as user_id, user_referrals.referral_id as referral_id").Find(&referrer).Error

	if rErr != nil {
		return nil, rErr
	}

	if referrer.UserId == nil {
		util.Log().Debug("no reward applicable. no referrer.")
		return nil, nil
	}
	referrerId := *referrer.UserId

	{ // debugging guard

		if selectedCashOrder.UserId != refereeId {
			return nil, errors.New("referreeIdFromSelected != refereeId")
		}

		if selectedCashOrder.Status != po.CashOrderStatusSuccess {
			return nil, errors.New("selectedCashOrder.Status != po.CashOrderStatusSuccess")
		}
	}

	effectiveCashInAmount := selectedCashOrder.EffectiveCashInAmount
	effectiveCashOutAmount := selectedCashOrder.EffectiveCashOutAmount
	orderType := selectedCashOrder.OrderType
	operationType := selectedCashOrder.OperationType
	selectedCashOrderId := selectedCashOrder.ID

	tetheredPromotion := TetheredRebatePromotion{
		SelectedCashOrderId:    selectedCashOrderId,
		ReferrerId:             referrerId,
		EffectiveCashInAmount:  effectiveCashInAmount,
		EffectiveCashOutAmount: effectiveCashOutAmount,
		OrderType:              orderType,
		OperationType:          operationType,
	}

	db.Create(&tetheredPromotion)

	return nil, nil
}

// TetheredRebatePromotion
type TetheredRebatePromotion struct {
	ReferrerId             int64
	SelectedCashOrderId    string
	EffectiveCashInAmount  int64
	EffectiveCashOutAmount int64
	OrderType              po.CashOrderOrderType
	OperationType          po.OperationType
	Rank                   int64
	Remarks                string
}

func (t *TetheredRebatePromotion) TableName() string {
	return "tethered_rebate_promotions"
}

// GetRewardsFor
func (s *Service) GetRewardsFor(ctx context.Context, referrerForm UserForm) error {
	panic("implement me")
	var referreeDbInfo po.User
	db := s.db.Where(`username`, referrerForm.Id)
	if err := db.Scopes(po.ByActiveNonStreamerUser).First(&referreeDbInfo).Error; err != nil {
		return err
	}

	//var referrer Referrer

	//rErr := db.Debug().Table("user_referrals").
	//	Joins("LEFT JOIN users ON users.id = user_referrals.referral_id").
	//	Where("user_referrals.referral_id = ?", referrerForm.Id).
	//	Select("users.id as user_id, user_referrals.referral_id as referral_id").Find(&referrer).Error
	//
	//if rErr != nil {
	//	return rErr
	//}

	return nil
}

const depositRankCountReadLimitHard = 10

func (s *Service) DepositRankCountReadLimit() int64 {
	if s.depositRankCountReadLimitF == nil {
		return depositRankCountReadLimitHard
	}
	a := s.depositRankCountReadLimitF()
	if a == 0 {
		a = depositRankCountReadLimitHard
	}
	return a
}

type _ = po.UserReferral
type _ = po.User
type _ = po.CashOrder

type Referrer struct {
	Id         int64  `gorm:"primarykey" gorm:"id"` // 主键ID
	UserId     *int64 `gorm:"column:user_id;"`
	ReferreeId int64  `gorm:"column:referral_id;"`
}

type RefereeCashOrder struct {
	ID                     string                `gorm:"primarykey" json:"id"` // 主键ID
	CreatedAt              time.Time             `gorm:"type:timestamp;"`      // 创建时间
	UpdatedAt              time.Time             `gorm:"type:timestamp;"`      // 更新时间
	UserId                 int64                 `gorm:"column:user_id"`
	TransactionId          *string               `gorm:"default:null;column:transaction_id"`
	CashMethodId           int64                 `gorm:"column:cash_method_id"`
	OrderType              po.CashOrderOrderType `gorm:"column:order_type"`
	Status                 po.CashOrderStatus    `gorm:"column:status"`
	Notes                  po.EncryptedStr       `gorm:"column:notes"`
	AppliedCashOutAmount   int64                 `gorm:"column:applied_cash_out_amount"`
	ActualCashOutAmount    int64                 `gorm:"column:actual_cash_out_amount"`
	BonusCashOutAmount     int64                 `gorm:"column:bonus_cash_out_amount"`
	EffectiveCashOutAmount int64                 `gorm:"column:effective_cash_out_amount"`
	AppliedCashInAmount    int64                 `gorm:"column:applied_cash_in_amount"`
	ActualCashInAmount     int64                 `gorm:"column:actual_cash_in_amount"`
	BonusCashInAmount      int64                 `gorm:"column:bonus_cash_in_amount"`
	EffectiveCashInAmount  int64                 `gorm:"column:effective_cash_in_amount"`
	BalanceBefore          int64                 `gorm:"column:balance_before"`
	WagerChange            int64                 `gorm:"column:wager_change"`
	Remark                 string                `gorm:"column:remark"`
	ManualClosedBy         int64                 `gorm:"column:manual_closed_by"`
	RequireReview          bool                  `gorm:"column:require_review"`
	ReviewStatus           int64                 `gorm:"column:review_status"`
	ApproveActionBy        int64                 `gorm:"column:approve_action_by"`
	ApprovedActionAt       time.Time             `gorm:"column:approved_action_at"`
	ApproveStatus          int64                 `gorm:"column:approve_status"`
	ConfirmActionBy        int64                 `gorm:"column:confirm_action_by"`
	ConfirmActionAt        time.Time             `gorm:"column:confirm_action_at"`
	ConfirmStatus          int64                 `gorm:"column:confirm_status"`
	IsManualOperation      bool                  `gorm:"column:is_manual_operation"`
	OperationType          po.OperationType      `gorm:"column:operation_type"`
	Ip                     string                `gorm:"column:ip;"`

	CallbackAt      time.Time `gorm:"column:callback_at;"`
	IsManualCashOut bool      `gorm:"column:is_manual_cash_out;"`
}

func (RefereeCashOrder) TableName() string {
	return "cash_orders"
}
