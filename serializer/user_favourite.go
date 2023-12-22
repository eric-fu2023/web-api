package serializer

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
)

type UserFavourite struct {
	GameId  int64 `json:"game_id,omitempty"`
	SportId int64 `json:"sport_id,omitempty"`
	Ref     int64 `json:"ref,omitempty"`
	*SubGameBrand
}

func BuildUserFavourite(a ploutos.UserFavourite) (b UserFavourite) {
	if a.SubGameBrand != nil {
		t := BuildSubGameBrand(*a.SubGameBrand)
		b.SubGameBrand = &t
	} else {
		b.GameId = a.GameId
		b.SportId = a.SportId
		b.Ref = a.Ref
	}
	return
}
