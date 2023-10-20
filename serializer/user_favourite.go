package serializer

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
)

type UserFavourite struct {
	GameId  int64         `json:"game_id"`
	SportId int64         `json:"sport_id,omitempty"`
	Game    *SubGameBrand `json:"game,omitempty"`
}

func BuildUserFavourite(a ploutos.UserFavourite) (b UserFavourite) {
	b = UserFavourite{
		GameId:  a.GameId,
		SportId: a.SportId,
	}
	if a.SubGameBrand != nil {
		t := BuildSubGameBrand(*a.SubGameBrand)
		b.Game = &t
	}
	return
}
