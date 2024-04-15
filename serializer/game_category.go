package serializer

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
)

type GameCategory struct {
	Id     int64        `json:"category_id"`
	Vendor []GameVendor `json:"vendor,omitempty"`
}

func BuildGameCategory(c *gin.Context, a ploutos.GameCategory, gameIds []int64) (b GameCategory) {
	b = GameCategory{
		Id: a.ID,
	}
	if len(a.GameVendor) > 0 {
		for i, v := range a.GameVendor {
			if len(gameIds) > i {
				t := BuildGameVendor(v, gameIds[i])
				b.Vendor = append(b.Vendor, t)
			}
		}
	}
	return
}
