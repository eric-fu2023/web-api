package model

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
)

type DollarJackpotBetReport struct {
	ploutos.DollarJackpotBetReport
	JackpotDraws  *ploutos.DollarJackpotDraw `gorm:"references:GameId;foreignKey:ID"`
}
