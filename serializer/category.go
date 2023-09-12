package serializer

import (
	"web-api/model"
)

type Category struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Icon string `json:"icon"`
}

func BuildCategory(a model.Category) (b Category) {
	b = Category{
		ID:   a.ID,
		Name: a.Name,
	}
	if a.Icon != "" {
		b.Icon = Url(a.Icon)
	}
	return
}
