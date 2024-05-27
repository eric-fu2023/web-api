package referral_alliance

import (
	"context"
	"encoding/json"
	"math/rand"
	"os"
	"sort"
	"time"
	"web-api/cache"
	"web-api/util"
)

const (
	mockRankingRangeRatePos = 1.1
	mockRankingRangeRateNeg = 0.95
)

type MockRanking struct {
	Nickname             string `json:"nickname"`
	Avatar               string `json:"avatar"`
	DefaultReferralCount int64  `json:"default_referral_count"`
	DefaultRewardAmount  int64  `json:"default_reward_amount"`
}

func fillAdditionalRankings(
	rankings cache.ReferralAllianceRankings,
) (retReward, retReferral []cache.ReferralAllianceRanking) {
	mockRankings := generateMockRankings()
	retReferral = fillAdditionalReferralRankings(rankings.ReferralRankings, mockRankings)
	retReward = fillAdditionalRewardRankings(rankings.RewardRankings, mockRankings)
	return retReward, retReferral
}

func fillAdditionalReferralRankings(cur, mock []cache.ReferralAllianceRanking) (ret []cache.ReferralAllianceRanking) {
	// sort mock according to the referral count
	sort.Slice(mock, func(i, j int) bool {
		return mock[i].ReferralCount > mock[j].ReferralCount
	})

	idxC, idxM := 0, 0
	for len(ret) < ReferralAllianceRankingsLimit && idxC < len(cur) && idxM < len(mock) {
		if cur[idxC].ReferralCount >= mock[idxM].ReferralCount {
			c := cur[idxC]
			ret = append(ret, c)
			idxC += 1
		} else {
			m := mock[idxM]
			ret = append(ret, m)
			idxM += 1
		}
	}

	if len(ret) < ReferralAllianceRankingsLimit {
		if idxC < len(cur) {
			remaining := ReferralAllianceRankingsLimit - len(ret)
			ret = append(ret, cur[idxC:idxC+remaining]...)
		} else {
			remaining := ReferralAllianceRankingsLimit - len(ret)
			ret = append(ret, mock[idxM:idxM+remaining]...)
		}
	}

	curRank := int64(1)
	for i := 0; i < len(ret); i++ {
		ret[i].Rank = curRank
		curRank += 1
	}
	return ret
}

func fillAdditionalRewardRankings(cur, mock []cache.ReferralAllianceRanking) (ret []cache.ReferralAllianceRanking) {
	// sort mock according to the referral count
	sort.Slice(mock, func(i, j int) bool {
		return mock[i].RewardAmount > mock[j].RewardAmount
	})

	idxC, idxM := 0, 0
	for len(ret) < ReferralAllianceRankingsLimit && idxC < len(cur) && idxM < len(mock) {
		if cur[idxC].RewardAmount >= mock[idxM].RewardAmount {
			c := cur[idxC]
			ret = append(ret, c)
			idxC += 1
		} else {
			m := mock[idxM]
			ret = append(ret, m)
			idxM += 1
		}
	}

	if len(ret) < ReferralAllianceRankingsLimit {
		if idxC < len(cur) {
			remaining := ReferralAllianceRankingsLimit - len(ret)
			ret = append(ret, cur[idxC:idxC+remaining]...)
		} else {
			remaining := ReferralAllianceRankingsLimit - len(ret)
			ret = append(ret, mock[idxM:idxM+remaining]...)
		}
	}

	curRank := int64(1)
	for i := 0; i < len(ret); i++ {
		ret[i].Rank = curRank
		curRank += 1
	}
	return ret
}

func generateMockRankings() []cache.ReferralAllianceRanking {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// json unmarshal
	var mockRankings []MockRanking
	jsonMockRankings := os.Getenv("REFERRAL_ALLIANCE_MOCK_RANKINGS")
	err := json.Unmarshal([]byte(jsonMockRankings), &mockRankings)
	if err != nil {
		util.GetLoggerEntry(context.Background()).Errorf("Err unmarshalling mock rankings err: %s", err.Error())
		return nil
	}

	var ret []cache.ReferralAllianceRanking
	for _, mockRanking := range mockRankings {
		rewardMaxValue := int64(float64(mockRanking.DefaultRewardAmount) * mockRankingRangeRatePos)
		rewardMinValue := int64(float64(mockRanking.DefaultRewardAmount) * mockRankingRangeRateNeg)
		ret = append(ret, cache.ReferralAllianceRanking{
			Nickname:      mockRanking.Nickname,
			Avatar:        mockRanking.Avatar,
			RewardAmount:  rewardMinValue + r.Int63n(rewardMaxValue-rewardMinValue+1),
			ReferralCount: mockRanking.DefaultReferralCount,
		})
	}

	return ret
}
