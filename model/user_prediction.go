package model

import (
	"errors"
	"time"
	"web-api/util"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"gorm.io/gorm"
)

type TempPrediction struct {
	AnalystId        int64   `json:"analyst_id"`
	PredictedMatches []Match `json:"predicted_matches"`
	PredictionTitle  string  `json:"prediction_title"`
	PredictionDesc   string  `json:"prediction_desc"`
}

type UserPrediction struct {
	ploutos.UserPrediction
	IsLocked bool `json:"is_locked"`
	Prediction TempPrediction `json:"prediction"`
}

type GetUserPredictionCond struct {
	DeviceId string
	UserId   int64
}

func GetUserPrediction(cond GetUserPredictionCond) ([]UserPrediction, error) {
	return GetUserPredictionWithDB(DB, cond)
}

func GetUserPredictionWithDB(tx *gorm.DB, cond GetUserPredictionCond) ([]UserPrediction, error) {
	if tx == nil {
		return nil, errors.New("tx is nil")
	}
	if cond.DeviceId == "" {
		return nil, errors.New("invalid uuid")
	}

	db := tx.Table(UserPrediction{}.TableName())

	db.Where("device_id = ?", cond.DeviceId)
	// if cond.UserId == 0 {
	// 	db.Where("user_id = 0")
	// }
	// db.Where("user_id = 0 OR user_id = ?", cond.UserId)

	now, err := util.NowGMT8()
	if err != nil {
		return nil, err
	}
	start := util.RoundDownTimeDay(now)
	end := util.RoundUpTimeDay(now)

	db.Where("created_at >= ?", start)
	db.Where("created_at < ?", end)

	var strategies []UserPrediction
	err = db.Find(&strategies).Error

	return strategies, err
}

func CreateUserPredictions(userId int64, deviceId string, predictionIds []int64) error {
	tx := DB.Begin()
	for _, predictionId := range predictionIds {
		err := CreateUserPredictionWithDB(tx, userId, deviceId, predictionId)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

func CreateUserPrediction(userId int64, deviceId string, predictionId int64) error {
	tx := DB.Begin()
	err := CreateUserPredictionWithDB(tx, userId, deviceId, predictionId)
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func CreateUserPredictionWithDB(tx *gorm.DB, userId int64, deviceId string, predictionId int64) error {
	if tx == nil {
		return errors.New("tx is nil")
	}

	// TODO : check if strategy exist

	obj := ploutos.UserPrediction{
		UserId:       userId,
		DeviceId:     deviceId,
		PredictionId: predictionId,
	}

	return tx.Create(&obj).Error

}

func MockGetUserPrediction(limit int) ([]UserPrediction, error) {
	val := []UserPrediction{
		{ 
			IsLocked: false,
			UserPrediction: ploutos.UserPrediction{
				DeviceId: "device101", UserId: 0, PredictionId: 1001, 
				BASE: ploutos.BASE{ID: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()},
			},
			Prediction: TempPrediction{AnalystId: 1, PredictionTitle: "WNBA 20-1 第一节，稳！", PredictionDesc: "法国第2场与荷兰踢了一场默契球，虽然缺和图拉姆都错失了绝佳进球机会，而拉比奥更是有机会也不射门，非要传给队友，这不得不让人怀疑双方是约定好的。目前2战过后积4分，排名第2，已经提前出线，但是获得第1可以避开德国、西班牙和葡萄牙这些上半区的强队，战意还是值得肯定的。不过目前法国国内的大选更牵动球员的心，赛前采访姆巴佩和图拉姆都号召年轻人去投票，显然这个场外因素对于球队影响巨大。波兰前2场全败，没有收获任何积分，上一场对阵荷兰莱万复出，反而拖慢了球队的进攻节奏，表现非常糟糕。上一场输给奥地利后，球队已经提前出局，本场比赛将为荣誉而战，这样一来，可以放下心理包袱，反而会超常发挥。两队有过3次交手，法国2胜1平，占据上风。最后，数据方面以法国让1.25球低奖和1.5球高奖起步，后市统一至1.75球中奖。法国实力远胜波兰，还有拿分意愿，战意值得肯定，不过队内球员分心国内大选，无法集中精力对待本场比赛，想要取得一场大胜存在困难。而波兰已经提前出局，轻装上阵的情况下，反而会发挥更佳，本场比赛也不会轻易认输。从市场的角度来看，初始让步力度强势，对于法国支持力度相当大，后市再次提升让步力度就太过夸张了，存在诱导嫌疑"},
		},
		{ 
			IsLocked: false,
			UserPrediction: ploutos.UserPrediction{
				DeviceId: "device101", UserId: 0, PredictionId: 1002, 
				BASE: ploutos.BASE{ID: 2, CreatedAt: time.Now(), UpdatedAt: time.Now()},
			},
			Prediction: TempPrediction{AnalystId: 2, PredictionTitle: "近20中14，欧洲杯精选，附半全场进球数玩法：法国vs波兰", PredictionDesc: "法国第2场与荷兰踢了一场默契球，虽然缺少了姆巴佩，但是阵容实力依然强劲，尤其是前锋线远胜荷兰，然而格列兹曼、登贝莱和图拉姆都错失了绝佳进球机会，而拉比奥更是有机会也不射门，非要传给队友，这不得不让人怀疑双方是约定好的。目前2战过后积4分，排名第2，已经提前出线，但是获得第1可以避开德国、西班牙和葡萄牙这些上半区的强队，战意还是值得肯定的。不过目前法国国内的大选更牵动球员的心，赛前采访姆巴佩和图拉姆都号召年轻人去投票，显然这个场外因素对于球队影响巨大。波兰前2场全败，没有收获任何积分，上一场对阵荷兰莱万复出，反而拖慢了球队的进攻节奏，表现非常糟糕。上一场输给奥地利后，球队已经提前出局，本场比赛将为荣誉而战，这样一来，可以放下心理包袱，反而会超常发挥。两队有过3次交手，法国2胜1平，占据上风。最后，数据方面以法国让1.25球低奖和1.5球高奖起步，后市统一至1.75球中奖。法国实力远胜波兰，还有拿分意愿，战意值得肯定，不过队内球员分心国内大选，无法集中精力对待本场比赛，想要取得一场大胜存在困难。而波兰已经提前出局，轻装上阵的情况下，反而会发挥更佳，本场比赛也不会轻易认输。从市场的角度来看，初始让步力度强势，对于法国支持力度相当大，后市再次提升让步力度就太过夸张了，存在诱导嫌疑"},
			
		},
		{ 
			IsLocked: false,
			UserPrediction: ploutos.UserPrediction{
				DeviceId: "device101", UserId: 0, PredictionId: 1003, 
				BASE: ploutos.BASE{ID: 3, CreatedAt: time.Now(), UpdatedAt: time.Now()},
			},
			Prediction: TempPrediction{AnalystId: 3, PredictionTitle: "欧洲杯8中7，昨日再次拿捏，今日法国vs比利时比赛分析，内附比分！！", PredictionDesc: "法国第2场与荷兰踢了然而格列兹曼、登贝莱和图拉姆都错失了绝佳进球机会，而拉比奥更是有机会也不射门，非要传给队友，这不得不让人怀疑双方是约定好的。目前2战过后积4分，排名第2，已经提前出线，但是获得第1可以避开德国、西班牙和葡萄牙这些上半区的强队，战意还是值得肯定的。不过目前法国国内的大选更牵动球员的心，赛前采访姆巴佩和图拉姆都号召年轻人去投票，显然这个场外因素对于球队影响巨大。波兰前2场全败，没有收获任何积分，上一场对阵荷兰莱万复出，反而拖慢了球队的进攻节奏，表现非常糟糕。上一场输给奥地利后，球队已经提前出局，本场比赛将为荣誉而战，这样一来，可以放下心理包袱，反而会超常发挥。两队有过3次交手，法国2胜1平，占据上风。最后，数据方面以法国让1.25球低奖和1.5球高奖起步，后市统一至1.75球中奖。法国实力远胜波兰，还有拿分意愿，战意值得肯定，不过队内球员分心国内大选，无法集中精力对待本场比赛，想要取得一场大胜存在困难。而波兰已经提前出局，轻装上阵的情况下，反而会发挥更佳，本场比赛也不会轻易认输。从市场的角度来看，初始让步力度强势，对于法国支持力度相当大，后市再次提升让步力度就太过夸张了，存在诱导嫌疑"},
		},
		{ 
			IsLocked: false,
			UserPrediction: ploutos.UserPrediction{
				DeviceId: "device101", UserId: 0, PredictionId: 1004, 
				BASE: ploutos.BASE{ID: 4, CreatedAt: time.Now(), UpdatedAt: time.Now()},
			},
			Prediction: TempPrediction{AnalystId: 4, PredictionTitle: "2串1 今天继续带来比赛分析，美职2串1！", PredictionDesc: "前法国国内的大选更牵动球员的心，赛前采访姆巴佩和图拉姆都号召年轻人去投票，显然这个场外因素对于球队影响巨大。波兰前2场全败，没有收获任何积分，上一场对阵荷兰莱万复出，反而拖慢了球队的进攻节奏，表现非常糟糕。上一场输给奥地利后，球队已经提前出局，本场比赛将为荣誉而战，这样一来，可以放下心理包袱，反而会超常发挥。两队有过3次交手，法国2胜1平，占据上风。最后，数据方面以法国让1.25球低奖和1.5球高奖起步，后市统一至1.75球中奖。法国实力远胜波兰，还有拿分意愿，战意值得肯定，不过队内球员分心国内大选，无法集中精力对待本场比赛，想要取得一场大胜存在困难。而波兰已经提前出局，轻装上阵的情况下，反而会发挥更佳，本场比赛也不会轻易认输。从市场的角度来看，初始让步力度强势，对于法国支持力度相当大，后市再次提升让步力度就太过夸张了，存在诱导嫌疑"},
		},
		{ 
			IsLocked: false,
			UserPrediction: ploutos.UserPrediction{
				DeviceId: "device101", UserId: 0, PredictionId: 1005, 
				BASE: ploutos.BASE{ID: 5, CreatedAt: time.Now(), UpdatedAt: time.Now()},
			},
			Prediction: TempPrediction{AnalystId: 5, PredictionTitle: "赢麻了赛分析，美职2串1！", PredictionDesc: "大选更内的大选更牵动球员的心，赛前采访姆巴佩和图拉姆都号召年轻人去投票，显然这个场外因素对于球队影响巨大。波兰前2场全败，没有收获任何积分，上一场对阵荷兰莱万复出，反而拖慢了球队的进攻节奏，表现非常糟糕。上一场输给奥地利后，球队已经提前出局，本场比赛将为荣誉而战，这样一来，可以放下心理包袱，反而会超常发挥。两队有过3次交手，法国2胜1平，占据上风。最后，数据方面以法国让1.25球低奖和1.5球高奖起步，后市统一至1.75球中奖。法国实力远胜波兰，还有拿分意愿，战意值得肯定，不过队内球员分心国内大选，无法集中精力对待本场比赛，想要取得一场大胜存在困难。而波兰已经提前出局，轻装上阵的情况下，反而会发挥更佳，本场比赛也不会轻易认输。从市场的角度来看，初始让步力度强势，对于法国支持力度相当大，后市再次提升让步力度就太过夸张了，存在诱导嫌疑"},
		},
	}

	for i := 0; i < limit; i++ {
		val[i].IsLocked = true
	}

	return val, nil 

}