package serializer

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"time"
)

type Voucher struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Image       string    `json:"image"`
	Type        int64     `json:"type"`
	StartAt     time.Time `json:"start_at"`
	EndAt       time.Time `json:"end_at"`
	Amount      float64   `json:"amount"`
	IsUsed      bool      `json:"is_used"`
}

func BuildVoucher(a ploutos.Voucher) (b Voucher) {
	b = Voucher{
		ID:          a.ID,
		Name:        a.Name,
		Description: a.Description,
		Image:       Url(a.Image),
		Type:        a.Type,
		StartAt:     a.StartAt,
		EndAt:       a.EndAt,
		Amount:      float64(a.Amount) / 100,
	}
	return
}
