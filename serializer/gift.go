package serializer

import (
	"sort"

	models "blgit.rfdev.tech/taya/ploutos-object"
)

type Gift struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	IsAnimated bool   `json:"is_animated"`
	// IconUrl      string `json:"icon_url"`
	// AnimationUrl string `json:"animation_url"`
	Price     float64 `json:"price"`
	Weightage int     `json:"weightage"`
}

func BuildGift(a []models.Gift) (b []Gift) {
	for _, gift := range a {
		b = append(b, Gift{
			ID:         gift.ID,
			Name:       gift.Name,
			IsAnimated: gift.IsAnimated,
			Price:      float64(gift.Price) / float64(100),
			Weightage:  gift.Weightage,
		})
	}

	sort.Slice(b, func(i, j int) bool {
		return b[i].Weightage > b[j].Weightage
	})
	return
}
