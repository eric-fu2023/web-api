package serializer

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
)

type UserTag struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

func BuildUserTag(a ploutos.UserTag) (b UserTag) {
	b = UserTag{
		Id:   a.ID,
		Name: a.Name,
	}
	return
}
