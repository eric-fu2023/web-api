package serializer

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"web-api/util"
)

type Wallet struct {
	Name         string  `json:"name"`
	GameCode     string  `json:"game_code"`
	Balance      float64 `json:"balance"`
	IsLastPlayed bool    `json:"is_last_played"`
}

func BuildWallet(a ploutos.GameVendorUser) (b Wallet) {
	b = Wallet{
		Balance:      util.MoneyFloat(a.Balance),
		IsLastPlayed: a.IsLastPlayed,
	}

	if a.GameVendor != nil {
		b.GameCode = a.GameVendor.GameCode
	}

	if a.GameVendor.GameVendorBrand != nil {
		b.Name = a.GameVendor.GameVendorBrand.Name
	}
	return
}
