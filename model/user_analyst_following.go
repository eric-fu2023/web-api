package model

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
)

type UserAnalystFollowing struct {
	ploutos.PredictionAnalystFollower

	Analyst Analyst `gorm:"foreignKey:AnalystId;references:ID"`
}