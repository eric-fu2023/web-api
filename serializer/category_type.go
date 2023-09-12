package serializer

import (
	"web-api/model"
)

type CategoryType struct {
	ID         int64      `json:"id"`
	Name       string     `json:"name"`
	Icon       string     `json:"icon"`
	Categories []Category `json:"categories,omitempty"`
}

func BuildCategoryType(a model.CategoryType) (b CategoryType) {
	b = CategoryType{
		ID:   a.ID,
		Name: a.Name,
	}
	if a.Icon != "" {
		b.Icon = Url(a.Icon)
	}
	if len(a.Categories) > 0 {
		for _, c := range a.Categories {
			b.Categories = append(b.Categories, BuildCategory(c))
		}
	}
	return
}
