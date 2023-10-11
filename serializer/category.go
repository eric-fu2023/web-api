package serializer

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
)

type Category struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Icon1 string `json:"icon1"`
	Icon2 string `json:"icon2"`
	Icon3 string `json:"icon3"`
}

func BuildCategory(a ploutos.Category) (b Category) {
	b = Category{
		ID:    a.ID,
		Name:  a.Name,
		Icon1: Url(a.Icon1),
		Icon2: Url(a.Icon2),
		Icon3: Url(a.Icon3),
	}
	return
}
