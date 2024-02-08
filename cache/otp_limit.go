package cache

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"os"
	"time"
)

const OtpCountKey = "otp_count:%s:%s" // otp_count:<date>:<grouping key>

// otpLimitScript has the following params
// KEYS[1]: Mobile key (mobile number value)
// KEYS[2]: Source key (uuid or ip)
// ARGV[1]: Mobile allowed limit
// ARGV[2]: Source allowed limit
var otpLimitScript = redis.NewScript(`
-- Key format: otp_count:<date>:<grouping key>
local mobileKey = KEYS[1]
local sourceKey = KEYS[2]

local mobileCount = tonumber(redis.call("GET", mobileKey) or 0)
local sourceCount = tonumber(redis.call("GET", sourceKey) or 0)

if mobileCount >= tonumber(ARGV[1]) or sourceCount >= tonumber(ARGV[2]) then
	return "false"
end

-- Increase limits
redis.call("INCR", mobileKey)
if mobileCount == 0 then
	-- If the key was just created let it expire after one day
    redis.call("EXPIRE", mobileKey, 86400)
end

redis.call("INCR", sourceKey)
if sourceCount == 0 then
	-- If the key was just created let it expire after one day
    redis.call("EXPIRE", sourceKey, 86400)
end

return "true"
`)

func IncreaseSendOtpLimit(mobile, ip, uuid string, datetime time.Time) (bool, error) {
	dateStr := datetime.Format("2006-01-02")
	sourceGroupingKey := fmt.Sprintf(OtpCountKey, dateStr, uuid)
	sourceLimit := os.Getenv("DAILY_OTP_LIMIT_UUID")
	if uuid == "" {
		sourceGroupingKey = fmt.Sprintf(OtpCountKey, dateStr, ip)
		sourceLimit = os.Getenv("DAILY_OTP_LIMIT_IP")
	}
	mobileLimit := os.Getenv("DAILY_OTP_LIMIT_MOBILE")
	mobileKey := fmt.Sprintf(OtpCountKey, dateStr, mobile)

	keys := []string{mobileKey, sourceGroupingKey}
	values := []interface{}{mobileLimit, sourceLimit}

	isWithinLimit, err := otpLimitScript.Run(context.Background(), RedisSessionClient, keys, values...).Bool()
	if err != nil {
		return false, err
	}

	return isWithinLimit, nil
}
