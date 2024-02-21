package serializer

import (
	"web-api/model"
)

type UserAchievement struct {
	AchievementId int64 `json:"achievement_id"`
	CreatedAt     int64 `json:"created_at"`
}

func BuildUserAchievements(achievements []model.UserAchievement) []UserAchievement {
	var resp []UserAchievement
	for _, a := range achievements {
		resp = append(resp, UserAchievement{
			AchievementId: a.AchievementId,
			CreatedAt:     a.CreatedAt.Unix(),
		})
	}
	return resp
}
