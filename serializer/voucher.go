package serializer

type Voucher struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Image       string `json:"image"`
	Type        int64  `json:"type"`
	StartAt     int64  `json:"start_at"`
	EndAt       int64  `json:"end_at"`
	Amount      int64  `json:"amount"`
	IsUsed      bool   `json:"is_used"`
}
