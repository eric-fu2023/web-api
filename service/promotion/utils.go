package promotion

import (
	"strconv"
	"time"
	"web-api/model"
)

// func getVipReward() int64 {

// }

func Today0am() time.Time {
	cfg, _ := model.GetAppConfigWithCache("timezone", "offset_seconds")
	offset, _ := strconv.Atoi(cfg)
	loc := time.FixedZone("", offset)
	date := time.Now().In(loc)
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, loc)
}
