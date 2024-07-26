package serializer

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
)

type SpinItem struct {
	ID        int64 `json:"id"`
	Name      string `json:"name"`
	PicSrc    string `json:"pic_src"`
	TextColor string `json:"text_color"`
	BgColor   string `json:"bg_color"`
}

func BuildSpinItem(a ploutos.SpinItem) (b SpinItem) {
	b = SpinItem{
		ID:        a.ID,
		Name:      a.Name,
		PicSrc:    a.PicSrc,
		TextColor: a.TextColor,
		BgColor:   a.BgColor,
	}
	return 
}


type SpinResult struct {
	ID        int64 `json:"id"`
}

func BuildSpinResult(a ploutos.SpinItem) (b SpinResult) {
	b = SpinResult{
		ID:        a.ID,
	}
	return 
}