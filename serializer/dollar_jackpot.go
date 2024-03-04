package serializer

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"web-api/util"
)

type DollarJackpot struct {
	Name           string   `json:"name"`
	Prize          float64  `json:"prize"`
	DescriptionWeb string   `json:"description_web"`
	DescriptionH5  string   `json:"description_h5"`
	Images         []string `json:"images,omitempty"`
}

func BuildDollarJackpot(c *gin.Context, a ploutos.DollarJackpot) (b DollarJackpot) {
	b = DollarJackpot{
		Name:           a.Name,
		Prize:          util.MoneyFloat(a.Prize),
		DescriptionWeb: Url(a.DescriptionWeb),
		DescriptionH5:  Url(a.DescriptionH5),
	}
	if a.Images != nil {
		var images []string
		if e := json.Unmarshal(a.Images, &images); e == nil {
			for i, img := range images {
				images[i] = Url(img)
			}
		}
		b.Images = images
	}
	return
}
