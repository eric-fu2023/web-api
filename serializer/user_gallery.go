package serializer

import (
	"web-api/model"
)

type UserGallery struct {
	ID     int64  `json:"id"`
	UserId int64  `json:"user_id"`
	Type   int64  `json:"type"`
	Src    string `json:"src"`
}

func BuildUserGallery(a model.UserGallery) (b UserGallery) {
	b = UserGallery{
		ID:     a.ID,
		UserId: a.UserId,
		Type:   a.Type,
		Src:    Url(a.Src),
	}
	return
}
