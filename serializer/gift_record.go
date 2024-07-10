package serializer

import (
	"time"

	models "blgit.rfdev.tech/taya/ploutos-object"
)

type GiftRecord struct {
	ID     int64 `json:"id"`
	UserId int64 `json:"user_id"`
	// GiftId       int64     `json:"gift_id"`
	GiftName     string    `json:"gift_name"`
	Quantity     int       `json:"quantity"`
	TotalPrice   int64     `json:"total_price"`
	LiveStreamId int64     `json:"live_stream_id"`
	GiftTime     time.Time `json:"gift_time"`
}

type PaginatedGiftRecord struct {
	TotalCount  int64        `json:"total_count"`
	TotalAmount float64      `json:"total_amount"`
	TotalWin    float64      `json:"total_win"`
	List        []GiftRecord `json:"list,omitempty"`
}

func BuildPaginatedGiftRecord(a []models.GiftRecord, total, amount, win int64) (b PaginatedGiftRecord) {

	b = PaginatedGiftRecord{
		TotalCount:  total,
		TotalAmount: float64(amount) / 100,
		TotalWin:    float64(win) / 100,
	}

	for _, giftRecord := range a {
		b.List = append(b.List, GiftRecord{
			ID:     giftRecord.ID,
			UserId: giftRecord.UserId,
			// GiftId:       giftRecord.GiftId,
			TotalPrice:   giftRecord.TotalPrice / 100,
			Quantity:     giftRecord.Quantity,
			LiveStreamId: giftRecord.LiveStreamId,
			GiftTime:     giftRecord.CreatedAt,
			GiftName:     giftRecord.Gift.Name,
		})
	}
	return
}
