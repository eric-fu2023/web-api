package model

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
)

type UserAnalystFollowing struct {
	//TODO need to update to ploutos.PredictionAnalystFollower... 
	ploutos.UserAnalystFollowing

	Analyst Analyst `gorm:"foreignKey:AnalystId;references:ID"`
}