package cache

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
	"time"
	"web-api/conf/consts"
)

var (
	otpExpiry           = 2 * time.Minute
	ErrInvalidOtpAction = errors.New("invalid sms otp action")

	isActionAvailable = map[string]bool{
		consts.SmsOtpActionLogin:                true,
		consts.SmsOtpActionDeleteUser:           true,
		consts.SmsOtpActionSetPassword:          true,
		consts.SmsOtpActionSetSecondaryPassword: true,
	}
)

func GetOtp(ctx context.Context, action, userKey string) (string, error) {
	key, err := buildOtpKey(action, userKey)
	if err != nil {
		return "", err
	}
	res := RedisClient.Get(ctx, key)
	if res.Err() != nil && res.Err() != redis.Nil {
		return "", res.Err()
	}
	return res.Val(), nil
}

func GetOtpByUserKeys(ctx context.Context, action string, userKeys []string) (otp string, err error) {
	for _, userKey := range userKeys {
		otp, err = GetOtp(ctx, action, userKey)
		if err != nil {
			return "", err
		}

		if otp != "" {
			break
		}
	}

	return otp, nil
}

func SetOtp(ctx context.Context, action, userKey, otp string) error {
	key, err := buildOtpKey(action, userKey)
	if err != nil {
		return err
	}
	res := RedisClient.Set(ctx, key, otp, otpExpiry)
	if res.Err() != nil {
		return res.Err()
	}
	return nil
}

func buildOtpKey(action, userKey string) (string, error) {
	if !isActionAvailable[action] {
		return "", ErrInvalidOtpAction
	}

	key := "otp:" + action + ":" + userKey
	return key, nil
}
