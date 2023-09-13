package serializer

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
)

type Match struct {
	ID           int64  `json:"id"`
	SportId      int64  `json:"sport_id"`
	LeagueId     int64  `json:"league_id"`
	LeagueNameEn string `json:"league_name_en"`
	LeagueNameCn string `json:"league_name_cn"`
	OpenTimeTs   int64  `json:"open_time_ts"`
	MatchStatus  int    `json:"match_status"`
	HomeId       int64  `json:"home_id"`
	HomeNameEn   string `json:"home_name_en"`
	HomeNameCn   string `json:"home_name_cn"`
	AwayId       int64  `json:"away_id"`
	AwayNameEn   string `json:"away_name_en"`
	AwayNameCn   string `json:"away_name_cn"`
	NamiId       int64  `json:"nami_id,omitempty"`
}

func BuildMatch(c *gin.Context, a ploutos.Match) (b Match) {
	b = Match{
		ID:           a.ID,
		SportId:      a.SportId,
		LeagueId:     a.LeagueId,
		LeagueNameEn: a.LeagueNameEn,
		LeagueNameCn: a.LeagueNameCn,
		OpenTimeTs:   a.OpenTime.Unix(),
		MatchStatus:  a.MatchStatus,
		HomeId:       a.HomeId,
		HomeNameEn:   a.HomeNameEn,
		HomeNameCn:   a.HomeNameCn,
		AwayId:       a.AwayId,
		AwayNameEn:   a.AwayNameEn,
		AwayNameCn:   a.AwayNameCn,
		NamiId:       a.NamiId,
	}

	return
}
