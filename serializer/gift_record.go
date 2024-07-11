package serializer

import (
	"strconv"

	models "blgit.rfdev.tech/taya/ploutos-object"
)

type GiftRecord struct {
	ID     string `json:"id"`
	UserId int64  `json:"user_id"`
	// GiftId       int64     `json:"gift_id"`
	GiftName      string `json:"gift_name"`
	Quantity      int    `json:"quantity"`
	TotalPrice    int64  `json:"total_price"`
	LiveStreamId  int64  `json:"live_stream_id"`
	GiftTime      int64  `json:"gift_time"`
	TransactionID string `json:"transaction_id"`
	StreamerName  string `json:"streamer_name"`
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

		dateTimeId := strconv.Itoa(int(giftRecord.CreatedAt.Unix())) + "T"
		uniqueId := dateTimeId + strconv.Itoa(int(giftRecord.ID))

		b.List = append(b.List, GiftRecord{
			ID:     uniqueId,
			UserId: giftRecord.UserId,
			// GiftId:       giftRecord.GiftId,
			TotalPrice:    giftRecord.TotalPrice / 100,
			Quantity:      giftRecord.Quantity,
			LiveStreamId:  giftRecord.LiveStreamId,
			GiftTime:      giftRecord.CreatedAt.Unix(),
			GiftName:      giftRecord.Gift.Name,
			StreamerName:  giftRecord.StreamerName,
			TransactionID: giftRecord.TransactionId,
		})
	}
	return
}
