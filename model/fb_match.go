package model

import (
	ploutosFB "blgit.rfdev.tech/taya/ploutos-object/fb/model"
)

type FbMatch struct {
	ploutosFB.FbMatch

	HomeTeam ploutosFB.FbTeam `gorm:"foriegnKey:HomeTeamId;references:TeamID"`
	AwayTeam ploutosFB.FbTeam `gorm:"foriegnKey:AwayTeamId;references:TeamID"`
	LeagueInfo ploutosFB.FbLeagues `gorm:"foreignKey:LeagueId;references:LeagueId"`
}