package referral_alliance

import (
	"math/rand"
	"time"
	"web-api/cache"
)

func fillAdditionalRankings(
	rankings cache.ReferralAllianceRankings,
) (retReward, retReferral []cache.ReferralAllianceRanking) {
	mockRankings := generateMockRankings()
	retReferral = fillAdditionalReferralRankings(rankings.ReferralRankings, mockRankings)
	retReward = fillAdditionalRewardRankings(rankings.RewardRankings, mockRankings)
	return retReward, retReferral
}

func fillAdditionalReferralRankings(cur, mock []cache.ReferralAllianceRanking) (ret []cache.ReferralAllianceRanking) {
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

	var ret []cache.ReferralAllianceRanking
	for _, mockRanking := range mockRewardRankings {
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

const (
	mockRankingRangeRatePos = 1.1
	mockRankingRangeRateNeg = 0.95
)

type MockRanking struct {
	Nickname             string
	Avatar               string
	DefaultReferralCount int64
	DefaultRewardAmount  int64
}

var (
	mockRewardRankings = []MockRanking{
		{
			Nickname:             "国米·布雷特",
			Avatar:               "/img/user/900/3e5408cf8ba2344e778c2e61035ccd_180_135.jpg",
			DefaultRewardAmount:  1309509,
			DefaultReferralCount: 135,
		},
		{

			Nickname:             "巴萨·诺尔",
			Avatar:               "/img/user/900/57c674ede3ecc1224beff71f6d6e5d_180_135.jpg",
			DefaultRewardAmount:  986578,
			DefaultReferralCount: 151,
		},
		{

			Nickname:             "斯巴达·贝吉塔",
			Avatar:               "/img/user/900/6631799bb681f8cf7f6ba9dda29547_180_135.jpg",
			DefaultRewardAmount:  503489,
			DefaultReferralCount: 98,
		},
		{

			Nickname:             "思考·仙女",
			Avatar:               "/img/user/900/694086c31abe683d5f7efdca7d1ba3_180_135.jpg",
			DefaultRewardAmount:  400380,
			DefaultReferralCount: 82,
		},
		{

			Nickname:             "煮饭·公牛",
			Avatar:               "/img/user/900/70afaf2a47802916ccc7cc9a5ff145_180_135.jpg",
			DefaultRewardAmount:  98823,
			DefaultReferralCount: 75,
		},
		{

			Nickname:             "敲木鱼·曹操",
			Avatar:               "/img/user/900/959267dd8549f78d3d3b4eeeeebcf2_180_135.jpg",
			DefaultRewardAmount:  45609,
			DefaultReferralCount: 34,
		},
		{

			Nickname:             "躲雨·荔枝",
			Avatar:               "/img/user/900/98949ee9b97c21787f3fcdf0ffbbd5_180_135.jpg",
			DefaultRewardAmount:  35515,
			DefaultReferralCount: 28,
		},
		{

			Nickname:             "坐电梯·仙女神",
			Avatar:               "/img/user/900/aca55cb864f180994847f157bf8354_180_135.jpg",
			DefaultRewardAmount:  15023,
			DefaultReferralCount: 25,
		},
		{

			Nickname:             "喝豆奶·编剧",
			Avatar:               "/img/user/900/afb5c20ba9ba21679108296b3462d8_180_135.jpg",
			DefaultRewardAmount:  6323,
			DefaultReferralCount: 12,
		},
		{

			Nickname:             "迫降·贝斯",
			Avatar:               "/img/user/900/b65a36cbfe158ab6cd71077fd62892_180_135.jpg",
			DefaultRewardAmount:  2503,
			DefaultReferralCount: 3,
		},
		{

			Nickname:             "富勒姆·布雷",
			Avatar:               "/img/user/900/c7b8179417ecadeab55e9036fb99e1_180_135.jpg",
			DefaultRewardAmount:  0,
			DefaultReferralCount: 122,
		},
		{

			Nickname:             "奥林匹亚科斯·威少",
			Avatar:               "/img/user/900/d949607c3cb2dd3148ed24bbfaefcf_180_135.jpg",
			DefaultRewardAmount:  0,
			DefaultReferralCount: 101,
		},
		{

			Nickname:             "掘金·希尔",
			Avatar:               "/img/user/900/e360dd9a5548b9205bd7374eeb0166_180_135.jpg",
			DefaultRewardAmount:  0,
			DefaultReferralCount: 87,
		},
	}
)
