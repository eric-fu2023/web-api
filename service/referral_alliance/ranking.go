package referral_alliance

import (
	"github.com/gin-gonic/gin"
	"web-api/model"
	"web-api/serializer"
	"web-api/util"
	"web-api/util/i18n"
)

type RankingsService struct{}

func (s *RankingsService) Get(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	resp := map[string]any{}

	if u, exists := c.Get("user"); exists {
		user, _ := u.(model.User)
		userInfo, err := s.GetUserInfo(user.ID)
		if err != nil {
			util.GetLoggerEntry(c).Errorf("GetUserInfo error: %s", err.Error())
			return serializer.GeneralErr(c, err), err
		}
		resp["user_info"] = userInfo
	}

	rewardRankings, err := s.GetRewardRankings()
	if err != nil {
		util.GetLoggerEntry(c).Errorf("GetRewardRankings error: %s", err.Error())
		return serializer.GeneralErr(c, err), err
	}
	resp["reward_rankings"] = rewardRankings

	referralRankings, err := s.GetReferralRankings()
	if err != nil {
		util.GetLoggerEntry(c).Errorf("GetReferralRankings error: %s", err.Error())
		return serializer.GeneralErr(c, err), err
	}
	resp["referral_rankings"] = referralRankings

	lastUpdateTime, err := s.GetLastUpdateTime()
	if err != nil {
		util.GetLoggerEntry(c).Errorf("GetLastUpdateTime error: %s", err.Error())
		return serializer.GeneralErr(c, err), err
	}
	resp["last_update_time"] = lastUpdateTime

	return serializer.Response{
		Data: resp,
		Msg:  i18n.T("success"),
	}, nil
}

func (s *RankingsService) GetRewardRankings() (any, error) {
	type RewardRanking struct {
		Rank         int64   `json:"rank"`
		Id           int64   `json:"id"`
		Nickname     string  `json:"nickname"`
		Avatar       string  `json:"avatar"`
		RewardAmount float64 `json:"reward_amount"`
	}

	resp := []RewardRanking{
		{
			Rank:         1,
			Id:           123,
			Nickname:     "potato1",
			Avatar:       "https://static.tayalive.com/img/user/224/avatar/224-avatar-20240404090035-PWauUp.jpg",
			RewardAmount: 10234.56,
		},
		{
			Rank:         2,
			Id:           534,
			Nickname:     "potato2",
			Avatar:       "https://static.tayalive.com/img/user/224/avatar/224-avatar-20240404090035-PWauUp.jpg",
			RewardAmount: 9234.22,
		},
		{
			Rank:         3,
			Id:           6745,
			Nickname:     "potato3",
			Avatar:       "https://static.tayalive.com/img/user/224/avatar/224-avatar-20240404090035-PWauUp.jpg",
			RewardAmount: 8234.33,
		},
		{
			Rank:         4,
			Id:           1256,
			Nickname:     "potato4",
			Avatar:       "https://static.tayalive.com/img/user/224/avatar/224-avatar-20240404090035-PWauUp.jpg",
			RewardAmount: 7234.44,
		},
		{
			Rank:         5,
			Id:           23,
			Nickname:     "potato",
			Avatar:       "https://static.tayalive.com/img/user/224/avatar/224-avatar-20240404090035-PWauUp.jpg",
			RewardAmount: 6234.55,
		},
		{
			Rank:         6,
			Id:           689,
			Nickname:     "potato6",
			Avatar:       "https://static.tayalive.com/img/user/224/avatar/224-avatar-20240404090035-PWauUp.jpg",
			RewardAmount: 5234.66,
		},
		{
			Rank:         7,
			Id:           9012,
			Nickname:     "potato7",
			Avatar:       "https://static.tayalive.com/img/user/224/avatar/224-avatar-20240404090035-PWauUp.jpg",
			RewardAmount: 4234.77,
		},
		{
			Rank:         8,
			Id:           5623,
			Nickname:     "potato8",
			Avatar:       "https://static.tayalive.com/img/user/224/avatar/224-avatar-20240404090035-PWauUp.jpg",
			RewardAmount: 3234.88,
		},
		{
			Rank:         9,
			Id:           843,
			Nickname:     "potato9",
			Avatar:       "https://static.tayalive.com/img/user/224/avatar/224-avatar-20240404090035-PWauUp.jpg",
			RewardAmount: 2234.99,
		},
		{
			Rank:         10,
			Id:           2367,
			Nickname:     "potato10",
			Avatar:       "https://static.tayalive.com/img/user/224/avatar/224-avatar-20240404090035-PWauUp.jpg",
			RewardAmount: 1234.12,
		},
	}

	return resp, nil
}

func (s *RankingsService) GetReferralRankings() (any, error) {
	type ReferralRanking struct {
		Rank          int64  `json:"rank"`
		Id            int64  `json:"id"`
		Nickname      string `json:"nickname"`
		Avatar        string `json:"avatar"`
		ReferralCount int64  `json:"referral_count"`
	}

	// Referral Rankings, ID should be random and unique, referral count should be in descending order, rank should be in ascending order
	resp := []ReferralRanking{
		{
			Rank:          1,
			Id:            843,
			Nickname:      "potato9",
			Avatar:        "https://static.tayalive.com/img/user/224/avatar/224-avatar-20240404090035-PWauUp.jpg",
			ReferralCount: 84,
		},
		{
			Rank:          2,
			Id:            534,
			Nickname:      "potato2",
			Avatar:        "https://static.tayalive.com/img/user/224/avatar/224-avatar-20240404090035-PWauUp.jpg",
			ReferralCount: 75,
		},
		{
			Rank:          3,
			Id:            689,
			Nickname:      "potato6",
			Avatar:        "https://static.tayalive.com/img/user/224/avatar/224-avatar-20240404090035-PWauUp.jpg",
			ReferralCount: 55,
		},
		{
			Rank:          4,
			Id:            1256,
			Nickname:      "potato4",
			Avatar:        "https://static.tayalive.com/img/user/224/avatar/224-avatar-20240404090035-PWauUp.jpg",
			ReferralCount: 40,
		},
		{
			Rank:          5,
			Id:            5623,
			Nickname:      "potato8",
			Avatar:        "https://static.tayalive.com/img/user/224/avatar/224-avatar-20240404090035-PWauUp.jpg",
			ReferralCount: 36,
		},
		{
			Rank:          6,
			Id:            123,
			Nickname:      "potato5",
			Avatar:        "https://static.tayalive.com/img/user/224/avatar/224-avatar-20240404090035-PWauUp.jpg",
			ReferralCount: 30,
		},
		{
			Rank:          7,
			Id:            9012,
			Nickname:      "potato7",
			Avatar:        "https://static.tayalive.com/img/user/224/avatar/224-avatar-20240404090035-PWauUp.jpg",
			ReferralCount: 14,
		},
		{
			Rank:          8,
			Id:            6745,
			Nickname:      "potato3",
			Avatar:        "https://static.tayalive.com/img/user/224/avatar/224-avatar-20240404090035-PWauUp.jpg",
			ReferralCount: 12,
		},

		{
			Rank:          9,
			Id:            123,
			Nickname:      "potato1",
			Avatar:        "https://static.tayalive.com/img/user/224/avatar/224-avatar-20240404090035-PWauUp.jpg",
			ReferralCount: 4,
		},
		{
			Rank:          10,
			Id:            2367,
			Nickname:      "potato10",
			Avatar:        "https://static.tayalive.com/img/user/224/avatar/224-avatar-20240404090035-PWauUp.jpg",
			ReferralCount: 3,
		},
	}

	return resp, nil
}

func (s *RankingsService) GetUserInfo(userId int64) (any, error) {
	type UserInfo struct {
		RewardRank    int64   `json:"reward_rank"`
		ReferralRank  int64   `json:"referral_rank"`
		Nickname      string  `json:"nickname"`
		Avatar        string  `json:"avatar"`
		RewardAmount  float64 `json:"reward_amount"`
		ReferralCount int64   `json:"referral_count"`
	}

	return UserInfo{
		RewardRank:    5,
		ReferralRank:  -1,
		Nickname:      "potato",
		Avatar:        "https://static.tayalive.com/img/user/224/avatar/224-avatar-20240404090035-PWauUp.jpg",
		RewardAmount:  6234.55,
		ReferralCount: 2,
	}, nil
}

func (s *RankingsService) GetLastUpdateTime() (any, error) {
	return 1715680935, nil
}
