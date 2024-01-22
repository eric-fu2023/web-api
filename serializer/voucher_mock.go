package serializer

import "time"

var VoucherMock = Voucher{
	ID:          1,
	Name:        "$100 v",
	Description: "earned",
	Image:       "xxx",
	StartAt:     time.Now(),
	EndAt:       time.Now().Add(24 * time.Hour),
	Type:        1,
	Amount:      100,
}

var VoucherMock2 = Voucher{
	ID:          1,
	Name:        "$100 v",
	Description: "earned",
	Image:       "xxx",
	StartAt:     time.Now().Add(-24 * time.Hour),
	EndAt:       time.Now(),
	Type:        1,
	Amount:      100,
}