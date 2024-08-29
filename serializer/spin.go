package serializer

import (
	"time"
	"web-api/conf/consts"
	"web-api/util/i18n"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
)

type Spin struct {
	ID              int64      `json:"id"`
	Name            string     `json:"name"`
	Description     string     `json:"description"`
	Button          string     `json:"button"`
	Counts          int        `json:"counts"`
	RemainingCounts int        `json:"remaining_counts"`
	PromotionId     int64      `json:"promotion_id"`
	SpinItems       []SpinItem `json:"items"`
}
type SpinItem struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	PicSrc    string `json:"pic_src"`
	TextColor string `json:"text_color"`
	BgColor   string `json:"bg_color"`
	IsWin     bool   `json:"is_win"`
}
type SpinHistory struct {
	SpinID         int64  `json:"spin_id"`
	SpinName       string `json:"spin_name"`
	SpinTime       int64  `json:"spin_time"`
	SpinResultId   int64  `json:"spin_result_id"`
	SpinResultName string `json:"spin_result_name"`
	SpinResultType string `json:"spin_result_type"`
	Redeemed       bool   `json:"redeemed"`
}
type SpinSqlHistory struct {
	SpinID         int64     `json:"spin_id"`
	SpinName       string    `json:"spin_name"`
	CreatedAt      time.Time `json:"created_at"`
	SpinResultId   int64     `json:"spin_result_id"`
	SpinResultName string    `json:"spin_result_name"`
	SpinResultType int64     `json:"spin_result_type"`
	Redeemed       bool      `json:"redeemed"`
}

func BuildSpin(spin ploutos.Spins, spin_items []ploutos.SpinItem, spin_result_counts int) (spin_resp Spin) {

	var spin_items_resp []SpinItem
	for _, item := range spin_items {
		spin_items_resp = append(spin_items_resp, BuildSpinItem(item))
	}
	spin_resp = Spin{
		ID:              spin.ID,
		Name:            spin.Name,
		Description:     spin.Description,
		Button:          spin.Button,
		Counts:          spin.Counts,
		RemainingCounts: spin.Counts - spin_result_counts,
		PromotionId:     spin.PromotionId,
		SpinItems:       spin_items_resp,
	}
	return
}

func BuildSpinItem(a ploutos.SpinItem) (b SpinItem) {
	text_color := ""
	bg_color := ""
	if len(a.TextColor) > 0 && a.TextColor[0] == '#' {
		text_color = a.TextColor[1:]
	}
	if len(a.BgColor) > 0 && a.BgColor[0] == '#' {
		bg_color = a.BgColor[1:]
	}
	b = SpinItem{
		ID:        a.ID,
		Name:      a.Name,
		PicSrc:    Url(a.PicSrc),
		TextColor: text_color,
		BgColor:   bg_color,
		IsWin:     a.IsWin,
	}
	return
}

type SpinResult struct {
	ID              int64 `json:"id"`
	RemainingCounts int   `json:"remaining_counts"`
}

func BuildSpinResult(a ploutos.SpinItem, remaining_counts int) (b SpinResult) {
	b = SpinResult{
		ID:              a.ID,
		RemainingCounts: remaining_counts,
	}
	return
}

func BuildSpinHistory(a []SpinSqlHistory, i18n i18n.I18n) (b []SpinHistory) {
	for _, result := range a {
		spin_result_type := i18n.T(consts.SpinResultType[result.SpinResultType])
		b = append(b, SpinHistory{
			SpinID:         result.SpinID,
			SpinName:       result.SpinName,
			SpinTime:       result.CreatedAt.Unix(),
			SpinResultId:   result.SpinResultId,
			SpinResultName: result.SpinResultName,
			SpinResultType: spin_result_type,
			Redeemed:       result.Redeemed,
		})
	}
	return
}
