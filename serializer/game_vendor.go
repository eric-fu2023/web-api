package serializer

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
)

type GameVendor struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	GameId  int64  `json:"game_id,omitempty"`
	WebIcon string `json:"web_icon,omitempty"`
	AppIcon string `json:"app_icon,omitempty"`
	Status  int64  `json:"status"`
}

func BuildGameVendor(a ploutos.GameVendor, gameId int64) (b GameVendor) {
	b = GameVendor{
		ID:      a.ID,
		Name:    a.Name,
		GameId:  gameId,
		WebIcon: Url(a.WebIcon),
		AppIcon: Url(a.AppIcon),
		Status:  a.Status,
	}
	return
}
