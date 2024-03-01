package model

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
)

type DollarJackpotDraw struct {
	ploutos.DollarJackpotDraw
	Total int64 `gorm:"-"`
}
