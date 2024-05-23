package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

const (
	referralAllianceRankingsKey     = "referral_alliance_rankings"
	referralAllianceUserRankingsKey = "referral_alliance_user_rankings:%d"
)

type ReferralAllianceRankings struct {
	RewardRankings   []ReferralAllianceRanking `json:"reward_rankings"`
	ReferralRankings []ReferralAllianceRanking `json:"referral_rankings"`
	LastUpdateTime   int64                     `json:"last_update_time,omitempty"`
}

type ReferralAllianceRanking struct {
	Id            int64  `json:"id"`
	Nickname      string `json:"nickname"`
	Avatar        string `json:"avatar"`
	RewardAmount  int64  `json:"reward_amount,omitempty"`
	ReferralCount int64  `json:"referral_count,omitempty"`
	Rank          int64  `json:"rank"`
}

func GetReferralAllianceRankings() (ReferralAllianceRankings, error) {
	res := RedisClient.Get(context.Background(), referralAllianceRankingsKey)
	if res.Err() != nil { // Return redis.Nil if record is not found
		return ReferralAllianceRankings{}, res.Err()
	}

	var rankings ReferralAllianceRankings
	err := json.Unmarshal([]byte(res.Val()), &rankings)
	if err != nil {
		return ReferralAllianceRankings{}, fmt.Errorf("unmarshal: %w", err)
	}

	return rankings, nil
}

func SetReferralAllianceRankings(rankings ReferralAllianceRankings, expiryDuration time.Duration) error {
	rankingsJson, err := json.Marshal(rankings)
	if err != nil {
		return err
	}

	res := RedisClient.Set(context.Background(), referralAllianceRankingsKey, rankingsJson, expiryDuration)
	if res.Err() != nil {
		return res.Err()
	}
	return nil
}

type UserReferralAllianceRanking struct {
	Id            int64   `json:"id"`
	Nickname      string  `json:"nickname"`
	Avatar        string  `json:"avatar"`
	RewardAmount  float64 `json:"reward_amount,omitempty"`
	ReferralCount int64   `json:"referral_count,omitempty"`
	RewardRank    int64   `json:"reward_rank"`
	ReferralRank  int64   `json:"referral_rank"`
}

func GetUserReferralAllianceRankings(userId int64) (UserReferralAllianceRanking, error) {
	key := fmt.Sprintf(referralAllianceUserRankingsKey, userId)
	res := RedisClient.Get(context.Background(), key)
	if res.Err() != nil { // Return redis.Nil if record is not found
		return UserReferralAllianceRanking{}, res.Err()
	}

	var ranking UserReferralAllianceRanking
	err := json.Unmarshal([]byte(res.Val()), &ranking)
	if err != nil {
		return UserReferralAllianceRanking{}, fmt.Errorf("unmarshal: %w", err)
	}

	return ranking, nil
}

func SetUserReferralAllianceRankings(userId int64, ranking UserReferralAllianceRanking, expiryDuration time.Duration) error {
	rankingJson, err := json.Marshal(ranking)
	if err != nil {
		return err
	}

	key := fmt.Sprintf(referralAllianceUserRankingsKey, userId)
	res := RedisClient.Set(context.Background(), key, rankingJson, expiryDuration)
	if res.Err() != nil {
		return res.Err()
	}
	return nil
}
