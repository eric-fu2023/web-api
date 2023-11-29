package model

import (
	"encoding/json"
	"errors"
	"time"

	models "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
)

const (
	BonusOperatorRatio int64 = iota
	Bddd
)

const (
	AdditionalRuleOperator int64 = iota
	Accc
)

// rule regulating bonus
type Bonus struct {
	models.Bonus
}

type BonusResult struct {
	Amount int64
	Err    error
}

// rule regulating order/user
type BonusRule struct {
	MinDeposit         int64
	BonusOperator      int64 // tiered, percentage,
	BonusOperatorParam json.RawMessage
	WagerMultiplier    int64 // may not be here
	ClaimLimit         int64 // may not be here
	VipLevel           int64
	ValidBefore        time.Time
	ValidAfter         time.Time
	RegistrationBefore time.Time
	RegistrationAfter  time.Time
	AdditionalRule     json.RawMessage
}

type additionalRule struct {
	operator int64
	params   json.RawMessage
}

type BonusHandler func(amount int64) *BonusResult

func (r BonusRule) ProcessMinDeposit() BonusHandler {
	return func(amount int64) *BonusResult {
		if amount < r.MinDeposit {
			return &BonusResult{
				Err: ErrLessThanMin,
			}
		} else {
			return &BonusResult{
				Amount: amount,
			}
		}
	}
}

func (r BonusRule) ProcessRegistrationTime(c *gin.Context) BonusHandler {
	user := c.MustGet("user").(User)

	return func(amount int64) *BonusResult {
		if user.SetupCompletedAt.Before(r.RegistrationBefore) {
			return &BonusResult{
				Err: ErrRegistrationTime,
			}
		} else if user.SetupCompletedAt.After(r.RegistrationAfter) {
			return &BonusResult{
				Err: ErrRegistrationTime,
			}
		} else {
			return &BonusResult{
				Amount: amount,
			}
		}
	}
}

func (r BonusRule) ProcessTime() BonusHandler {
	now := time.Now()
	return func(amount int64) *BonusResult {
		if now.Before(r.ValidBefore) {
			return &BonusResult{
				Err: ErrNow,
			}
		} else if now.After(r.ValidAfter) {
			return &BonusResult{
				Err: ErrNow,
			}
		} else {
			return &BonusResult{
				Amount: amount,
			}
		}
	}
}

func (r BonusRule) BonusOperatorHandler() BonusHandler {
	switch r.BonusOperator {
	case BonusOperatorRatio:
		params := []float64{}
		b, _ := r.BonusOperatorParam.MarshalJSON()
		json.Unmarshal(b, &params)
		ratio := params[0]
		upperLimit := params[1]
		return func(amount int64) *BonusResult {
			return &BonusResult{
				Amount: int64(min(float64(amount)*ratio, upperLimit)),
			}
		}
	default:
		return func(amount int64) *BonusResult {
			return &BonusResult{
				Err: ErrInvalidOperator,
			}
		}
	}
}

func (r BonusRule) VipHandler() BonusHandler {
	var vipLevel int64
	return func(amount int64) *BonusResult {
		if vipLevel < r.VipLevel {
			return &BonusResult{
				Err: ErrVip,
			}
		} else {
			return &BonusResult{
				Amount: amount,
			}
		}
	}
}

var (
	ErrLessThanMin      = errors.New("less_than_min_deposit")
	ErrInvalidOperator  = errors.New("invalid_operator")
	ErrVip              = errors.New("vip_level_violation")
	ErrNow              = errors.New("not_valid_now")
	ErrRegistrationTime = errors.New("not_valid_registration_time")
)

func (r *BonusResult) Bind(h BonusHandler) *BonusResult {
	if r.Err != nil {
		return r
	}
	return h(r.Amount)
}

func (r *BonusResult) Handle(c *gin.Context, rule BonusRule) *BonusResult {
	return r.Bind(rule.ProcessMinDeposit())
}
