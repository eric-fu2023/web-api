package serializer

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"web-api/util"
)

type DollarJackpot struct {
	Name        string   `json:"name"`
	Prize       float64  `json:"prize"`
	Description string   `json:"description"`
	Images      []string `json:"images,omitempty"`
}

func BuildDollarJackpot(c *gin.Context, a ploutos.DollarJackpot) (b DollarJackpot) {
	b = DollarJackpot{
		Name:        a.Name,
		Prize:       util.MoneyFloat(a.Prize),
		Description: Url(a.Description),
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
