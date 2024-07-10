package serializer

import (
	"time"

	models "blgit.rfdev.tech/taya/ploutos-object"
)

type GiftRecord struct {
	ID           int64     `json:"id"`
	UserId       int64     `json:"user_id"`
	GiftId       int64     `json:"gift_id"`
	Quantity     int       `json:"quantity"`
	TotalPrice   int64     `json:"total_price"`
	LiveStreamId int64     `json:"live_stream_id"`
	GiftTime     time.Time `json:"gift_time"`
}

func BuildGiftRecords(a []models.GiftRecord) (b []GiftRecord) {
	for _, giftRecord := range a {
		b = append(b, GiftRecord{
			ID:           giftRecord.ID,
			UserId:       giftRecord.UserId,
			GiftId:       giftRecord.GiftId,
			TotalPrice:   giftRecord.TotalPrice / 100,
			Quantity:     giftRecord.Quantity,
			LiveStreamId: giftRecord.LiveStreamId,
			GiftTime:     giftRecord.CreatedAt,
		})
	}
	return
}
