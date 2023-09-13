package serializer

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
)

type Category struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Icon string `json:"icon"`
}

func BuildCategory(a ploutos.Category) (b Category) {
	b = Category{
		ID:   a.ID,
		Name: a.Name,
	}
	if a.Icon != "" {
		b.Icon = Url(a.Icon)
	}
	return
}
