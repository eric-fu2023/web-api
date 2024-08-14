package serializer

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
)

type Spin struct {
	ID              int64      `json:"id"`
	Name            string     `json:"name"`
	Description     string     `json:"description"`
	Button          string     `json:"button"`
	Counts          int        `json:"counts"`
	RemainingCounts int        `json:"remaining_counts"`
	PromotionId     int8       `json:"promotion_id"`
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
	b = SpinItem{
		ID:        a.ID,
		Name:      a.Name,
		PicSrc:    a.PicSrc,
		TextColor: a.TextColor,
		BgColor:   a.BgColor,
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
