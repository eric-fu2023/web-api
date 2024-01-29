package serializer

import (
	"encoding/json"
	"web-api/model"

	models "blgit.rfdev.tech/taya/ploutos-object"
)

type Voucher struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Image       string  `json:"image"`
	Type        int64   `json:"type"`
	StartAt     int64   `json:"start_at"`
	EndAt       int64   `json:"end_at"`
	Amount      float64 `json:"amount"`
	Status      int     `json:"status"`
}

func BuildVoucher(a models.Voucher, platform string) (b Voucher) {
	raw := json.RawMessage(a.Image)
	m := make(map[string]string)
	json.Unmarshal(raw, &m)
	image := m[platform]
	if len(image) == 0 {
		image = m["h5"]
	}
	b = Voucher{
		ID:          a.ID,
		Name:        a.Name,
		Description: a.Description,
		Image:       Url(a.Image),
		Type:        a.PromotionType,
		StartAt:     a.StartAt.Unix(),
		EndAt:       a.EndAt.Unix(),
		Amount:      float64(a.Amount) / 100,
		Status:      a.Status,
	}
	return
}

func BuildVoucherFromTemplate(a models.VoucherTemplate, amount int64, platform string) (b Voucher) {
	raw := json.RawMessage(a.Image)
	m := make(map[string]string)
	json.Unmarshal(raw, &m)
	image := m[platform]
	if len(image) == 0 {
		image = m["h5"]
	}
	displayAmount := amount / 100
	name := model.AmountReplace(a.Name, displayAmount)
	b = Voucher{
		ID:          a.ID,
		Name:        name,
		Description: a.Description,
		Image:       Url(a.Image),
		Type:        a.PromotionType,
		StartAt:     a.StartAt.Unix(),
		EndAt:       a.EndAt.Unix(),
		Amount:      0,
		Status:      0,
	}
	return
}
