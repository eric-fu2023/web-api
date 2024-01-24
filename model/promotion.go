package model

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	models "blgit.rfdev.tech/taya/ploutos-object"
	"gorm.io/gorm"
)

type Promotion struct {
	models.Promotion
}

func (p Promotion) GetEligibilityDetail() (ret EligibilityDetails) {
	_ = json.Unmarshal(p.EligibilityDetails, &ret)
	return
}

type EligibilityDetails struct {
	Type       string
	Repeatable Filter
}

type ReferenceValue string

const (
	VoucherCreatedAt ReferenceValue = "voucher_created_at"
	TopUpAt          ReferenceValue = "top_up_at"
)

type Operator string

const (
	Lt  Operator = "lt"
	Gt  Operator = "gt"
	Lte Operator = "lte"
	Gte Operator = "gte"
	Eq  Operator = "eq"
)

var operatorMap = map[Operator]string{
	Lt:  "<",
	Gt:  ">",
	Lte: "<=",
	Gte: ">=",
	Eq:  "=",
}

type ValueType string

const (
	TimeOfWeek     ValueType = "time_of_week"
	LastTimeOfWeek ValueType = "last_time_of_week"
	Absolute       ValueType = "absolute"
	Relative       ValueType = "relative"
)

type Filter struct {
	ReferenceValue ReferenceValue `json:"reference_value"`
	Operator       Operator       `json:"operator"`
	Value          string         `json:"value"`
	ValueType      ValueType      `json:"value_type"`
}

func (f Filter) Scope(field string) func(tx *gorm.DB) *gorm.DB {
	return func(tx *gorm.DB) *gorm.DB {
		var t time.Time
		switch f.ValueType {
		case Absolute:
			t, _ = time.Parse(time.DateTime, f.Value)
		case Relative:
			d, _ := time.ParseDuration(f.Value)
			t = time.Now().Add(d)
		case TimeOfWeek: // add tz for weekday
			now := time.Now()
			arr := strings.Split(f.Value, " ")
			weekday, _ := strconv.Atoi(arr[1])
			ts, _ := time.Parse(time.TimeOnly, arr[0])
			day := now.Day() + (int(now.Weekday()) - weekday)
			t = time.Date(now.Year(), now.Month(), day, ts.Hour(), ts.Minute(), ts.Second(), 0, time.UTC)
		case LastTimeOfWeek:
			now := time.Now()
			arr := strings.Split(f.Value, " ")
			weekday, _ := strconv.Atoi(arr[1])
			ts, _ := time.Parse(time.TimeOnly, arr[0])
			day := now.Day() + (int(now.Weekday()) - weekday)
			thisWeekTime := time.Date(now.Year(), now.Month(), day, ts.Hour(), ts.Minute(), ts.Second(), 0, time.UTC)
			if now.Before(thisWeekTime) {
				t = thisWeekTime.AddDate(0, 0, -7)
			} else {
				t = thisWeekTime
			}
		default:
			return tx
		}
		return tx.Where(fmt.Sprintf("%s %s ?", field, parseOperator(f.Operator)), t)
	}
}

func (f Filter) Condition(referenceValue int64) bool {
	v, _ := strconv.Atoi(f.Value)
	value := int64(v)
	switch f.Operator {
	case Lt:
		return value < referenceValue
	case Gt:
		return value > referenceValue
	case Lte:
		return value <= referenceValue
	case Gte:
		return value >= referenceValue
	case Eq:
		return value == referenceValue
	default:
		return false
	}
}

func parseOperator(o Operator) string {
	return operatorMap[o]
}
