package serializer

import "time"

var VoucherMock = Voucher{
	ID:          1,
	Name:        "$100 v",
	Description: "earned",
	Image:       "xxx",
	StartAt:     time.Now().Unix(),
	EndAt:       time.Now().Add(24 * time.Hour).Unix(),
	Type:        1,
	Amount:      100,
	IsUsed:      true,
}

var VoucherMock2 = Voucher{
	ID:          1,
	Name:        "$100 v",
	Description: "earned",
	Image:       "xxx",
	StartAt:     time.Now().Add(-24 * time.Hour).Unix(),
	EndAt:       time.Now().Unix(),
	Type:        1,
	Amount:      100,
	IsUsed:      false,
}
