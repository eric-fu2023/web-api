package serializer

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
)

type GameCategory struct {
	Id     int64       `json:"category_id"`
	Vendor *GameVendor `json:"vendor,omitempty"`
}

func BuildGameCategory(c *gin.Context, a ploutos.GameCategory, gameId int64) (b GameCategory) {
	b = GameCategory{
		Id: a.ID,
	}
	if a.GameVendor != nil {
		t := BuildGameVendor(*a.GameVendor, gameId)
		b.Vendor = &t
	}
	return
}
