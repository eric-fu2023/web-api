package referral_alliance

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"strconv"
	"time"
	"web-api/cache"
	"web-api/model"
	"web-api/serializer"
	"web-api/util"
	"web-api/util/i18n"
)

const (
	ReferralAllianceRankingsLimit = 10
)

type RankingsService struct{}

func (s *RankingsService) Get(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	resp := map[string]any{}

	rankings, err := s.GetRankings(c)
	if err != nil {
		util.GetLoggerEntry(c).Errorf("GetRankings error: %s", err.Error())
		return serializer.GeneralErr(c, err), err
	}
	resp["reward_rankings"] = rankings.RewardRankings
	resp["referral_rankings"] = rankings.ReferralRankings
	resp["last_update_time"] = rankings.LastUpdateTime

	if u, exists := c.Get("user"); exists {
		user, _ := u.(model.User)
		userInfo, err := s.GetUserInfo(c, user, rankings)
		if err != nil {
			util.GetLoggerEntry(c).Errorf("GetUserInfo error: %s", err.Error())
			return serializer.GeneralErr(c, err), err
		}
		resp["user_info"] = userInfo
	}

	return serializer.Response{
		Data: resp,
		Msg:  i18n.T("success"),
	}, nil
}

func (s *RankingsService) GetRankings(ctx context.Context) (serializer.ReferralAllianceRankings, error) {
	now := time.Now()

	rankingsCache, err := s.GetRankingsFromCache(ctx)
	if err != nil && !errors.Is(err, redis.Nil) {
		util.GetLoggerEntry(ctx).Errorf("GetRankingsFromCache error: %s", err.Error())
	}

	if err == nil {
		return s.buildReferralAllianceRankingsResponseFromCache(rankingsCache), nil
	}

	rewardRankingsDB, referralRankingsDB, err := s.GetRankingsFromDB(ctx)
	if err != nil {
		util.GetLoggerEntry(ctx).Errorf("GetRankingsFromDB error: %s", err.Error())
		return serializer.ReferralAllianceRankings{}, err
	}

	rankingsCache = s.buildReferralAllianceRankingsCacheFromDb(rewardRankingsDB, referralRankingsDB, now)
	err = s.SetRankingsToCache(ctx, rankingsCache, now)
	if err != nil {
		util.GetLoggerEntry(ctx).Errorf("SetRankingsToCache error: %s", err.Error())
		// Just log error
	}

	return s.buildReferralAllianceRankingsResponseFromCache(rankingsCache), nil
}

func (s *RankingsService) GetRankingsFromCache(ctx context.Context) (cache.ReferralAllianceRankings, error) {
	rankingsCache, err := cache.GetReferralAllianceRankings()
	if err != nil && !errors.Is(err, redis.Nil) {
		util.GetLoggerEntry(ctx).Errorf("GetReferralAllianceRankings error: %s", err.Error())
	}
	if err != nil {
		return cache.ReferralAllianceRankings{}, err
	}

	return rankingsCache, nil
}

func (s *RankingsService) SetRankingsToCache(ctx context.Context, rankings cache.ReferralAllianceRankings, now time.Time) error {
	nowTz, err := s.getTzAdjustedTime(now)
	if err != nil {
		util.GetLoggerEntry(ctx).Errorf("getTzAdjustedTime error: %s", err.Error())
		return err
	}

	expiryDuration := s.next5PM(nowTz).Sub(nowTz)
	if expiryDuration < time.Second {
		expiryDuration = time.Second
	}

	err = cache.SetReferralAllianceRankings(rankings, expiryDuration)
	if err != nil {
		util.GetLoggerEntry(ctx).Errorf("GetReferralAllianceRankings error: %s", err.Error())
		return err
	}

	return nil
}

func (s *RankingsService) GetRankingsFromDB(ctx context.Context) (
	rewardRankings []model.ReferralAllianceRankingInfo,
	referralRankings []model.ReferralAllianceRankingInfo,
	err error,
) {
	rewardRankings, err = model.GetTopReferralAllianceRewardRankings(ReferralAllianceRankingsLimit)
	if err != nil {
		util.GetLoggerEntry(ctx).Errorf("GetTopReferralAllianceRewardRankings error: %s", err.Error())
		return nil, nil, err
	}

	referralRankings, err = model.GetTopReferralAllianceReferralRankings(ReferralAllianceRankingsLimit)
	if err != nil {
		util.GetLoggerEntry(ctx).Errorf("GetTopReferralAllianceReferralRankings error: %s", err.Error())
		return nil, nil, err
	}

	return rewardRankings, referralRankings, nil
}

func (s *RankingsService) GetUserInfo(ctx context.Context, user model.User, rankings serializer.ReferralAllianceRankings) (serializer.ReferralAllianceRankingUserRanking, error) {
	now := time.Now()

	userRankingCache, err := cache.GetUserReferralAllianceRankings(user.ID)
	if err != nil && !errors.Is(err, redis.Nil) {
		util.GetLoggerEntry(ctx).Errorf("GetRankingsFromCache error: %s", err.Error())
		return serializer.ReferralAllianceRankingUserRanking{}, err
	}

	// Return if found in cache
	if err == nil {
		return serializer.ReferralAllianceRankingUserRanking{
			ReferralAllianceRankingUserDetails: serializer.ReferralAllianceRankingUserDetails{
				Id:            userRankingCache.Id,
				Nickname:      userRankingCache.Nickname,
				Avatar:        userRankingCache.Avatar,
				RewardAmount:  userRankingCache.RewardAmount,
				ReferralCount: userRankingCache.ReferralCount,
			},
			RewardRank:   userRankingCache.RewardRank,
			ReferralRank: userRankingCache.ReferralRank,
		}, nil
	}

	// Get from DB if cache miss
	userRanking, err := model.GetUserReferralAllianceRankingInfo(user.ID)
	if err != nil {
		util.GetLoggerEntry(ctx).Errorf("GetUserReferralAllianceRankingInfo error: %s", err.Error())
		return serializer.ReferralAllianceRankingUserRanking{}, err
	}

	ret := serializer.ReferralAllianceRankingUserRanking{
		ReferralAllianceRankingUserDetails: serializer.ReferralAllianceRankingUserDetails{
			Id:            user.ID,
			Nickname:      user.Nickname,
			Avatar:        serializer.Url(user.Avatar),
			RewardAmount:  float64(userRanking.TotalClaimable / 100),
			ReferralCount: userRanking.ReferralCount,
		},
		RewardRank:   -1,
		ReferralRank: -1,
	}

	for _, ranking := range rankings.RewardRankings {
		if ranking.Id == user.ID {
			ret.RewardRank = ranking.Rank
			break
		}
	}
	for _, ranking := range rankings.ReferralRankings {
		if ranking.Id == user.ID {
			ret.RewardRank = ranking.Rank
			break
		}
	}

	// Set to cache
	err = s.SetUserRankingToCache(ctx, user, s.buildUserReferralAllianceRankingsCache(ret), now)
	if err != nil {
		util.GetLoggerEntry(ctx).Errorf("SetUserRankingToCache error: %s", err.Error())
		return serializer.ReferralAllianceRankingUserRanking{}, err
	}

	return ret, nil
}

func (s *RankingsService) SetUserRankingToCache(ctx context.Context, user model.User, ranking cache.UserReferralAllianceRanking, now time.Time) error {
	nowTz, err := s.getTzAdjustedTime(now)
	if err != nil {
		util.GetLoggerEntry(ctx).Errorf("getTzAdjustedTime error: %s", err.Error())
		return err
	}

	expiryDuration := s.next5PM(nowTz).Sub(nowTz)
	if expiryDuration < time.Second {
		expiryDuration = time.Second
	}

	err = cache.SetUserReferralAllianceRankings(user.ID, ranking, expiryDuration)
	if err != nil && errors.Is(err, redis.Nil) {
		util.GetLoggerEntry(ctx).Errorf("GetReferralAllianceRankings error: %s", err.Error())
		return err
	}

	return nil
}

func (s *RankingsService) getTzAdjustedTime(t time.Time) (time.Time, error) {
	tzOffsetStr, err := model.GetAppConfigWithCache("timezone", "offset_seconds")
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get tz offset config: %w", err)
	}
	tzOffset, err := strconv.Atoi(tzOffsetStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse tz offset config: %w", err)
	}

	return t.In(time.FixedZone("", tzOffset)), nil
}

func (s *RankingsService) next5PM(t time.Time) time.Time {
	// Create a new time at 5pm today
	today5PM := time.Date(t.Year(), t.Month(), t.Day(), 17, 0, 0, 0, t.Location())

	// If the current time is before or exactly 5pm, return today's 5pm
	if !t.After(today5PM) {
		return today5PM
	}

	// Otherwise, return 5pm of the next day
	return today5PM.Add(24 * time.Hour)
}

func (s *RankingsService) buildReferralAllianceRankingsResponseFromCache(rankingsCache cache.ReferralAllianceRankings) serializer.ReferralAllianceRankings {
	var rewardRankings []serializer.ReferralAllianceRanking
	for _, rankingCache := range rankingsCache.RewardRankings {
		rewardRankings = append(rewardRankings, serializer.ReferralAllianceRanking{
			ReferralAllianceRankingUserDetails: serializer.ReferralAllianceRankingUserDetails{
				Id:           rankingCache.Id,
				Nickname:     rankingCache.Nickname,
				Avatar:       serializer.Url(rankingCache.Avatar),
				RewardAmount: float64(rankingCache.RewardAmount / 100),
			},
			Rank: rankingCache.Rank,
		})
	}

	var referralRankings []serializer.ReferralAllianceRanking
	for _, rankingCache := range rankingsCache.ReferralRankings {
		referralRankings = append(referralRankings, serializer.ReferralAllianceRanking{
			ReferralAllianceRankingUserDetails: serializer.ReferralAllianceRankingUserDetails{
				Id:            rankingCache.Id,
				Nickname:      rankingCache.Nickname,
				Avatar:        serializer.Url(rankingCache.Avatar),
				ReferralCount: rankingCache.ReferralCount,
			},
			Rank: rankingCache.Rank,
		})
	}

	ret := serializer.ReferralAllianceRankings{
		RewardRankings:   rewardRankings,
		ReferralRankings: referralRankings,
		LastUpdateTime:   rankingsCache.LastUpdateTime,
	}

	return ret
}

func (s *RankingsService) buildReferralAllianceRankingsCacheFromDb(
	rewardRankingsDB,
	referralRankingsDB []model.ReferralAllianceRankingInfo,
	now time.Time,
) cache.ReferralAllianceRankings {
	var rewardRankingsCache []cache.ReferralAllianceRanking
	curRank := int64(1)
	for _, rankingDB := range rewardRankingsDB {
		if rankingDB.Referrer == nil {
			continue
		}
		rewardRankingsCache = append(rewardRankingsCache, cache.ReferralAllianceRanking{
			Id:           rankingDB.Referrer.ID,
			Nickname:     rankingDB.Referrer.Nickname,
			Avatar:       rankingDB.Referrer.Avatar,
			RewardAmount: rankingDB.TotalClaimable,
			Rank:         curRank,
		})
		curRank += 1
	}

	var referralRankingsCache []cache.ReferralAllianceRanking
	curRank = int64(1)
	for _, rankingDB := range referralRankingsDB {
		if rankingDB.Referrer == nil {
			continue
		}
		referralRankingsCache = append(referralRankingsCache, cache.ReferralAllianceRanking{
			Id:            rankingDB.Referrer.ID,
			Nickname:      rankingDB.Referrer.Nickname,
			Avatar:        rankingDB.Referrer.Avatar,
			ReferralCount: rankingDB.ReferralCount,
			Rank:          curRank,
		})
		curRank += 1
	}

	ret := cache.ReferralAllianceRankings{
		RewardRankings:   rewardRankingsCache,
		ReferralRankings: referralRankingsCache,
		LastUpdateTime:   now.Unix(),
	}
	return ret
}

func (s *RankingsService) buildUserReferralAllianceRankingsCache(userRanking serializer.ReferralAllianceRankingUserRanking) cache.UserReferralAllianceRanking {
	return cache.UserReferralAllianceRanking{
		Id:            userRanking.Id,
		Nickname:      userRanking.Nickname,
		Avatar:        userRanking.Avatar,
		RewardAmount:  userRanking.RewardAmount,
		ReferralCount: userRanking.ReferralCount,
		RewardRank:    userRanking.RewardRank,
		ReferralRank:  userRanking.ReferralRank,
	}
}
