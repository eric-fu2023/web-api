package cache

import (
	"errors"
	"fmt"
	"github.com/chenyahui/gin-cache/persist"
	"time"
)

const (
	appConfigKey        = "app_config:%d:%d:%s" // app_config:<brand>:<platform>:<key>
	appConfigExpiryTime = 10 * time.Minute
)

func GetAppConfig(brandId int, platform int64, key string) (cf map[string]map[string]string, err error) {
	key = fmt.Sprintf(appConfigKey, brandId, platform, key)

	err = RedisStore.Get(key, &cf)
	if err != nil && errors.Is(err, persist.ErrCacheMiss) {
		return nil, ErrCacheMiss
	}
	if err != nil {
		return nil, err
	}

	return cf, nil
}

func SetAppConfig(brandId int, platform int64, key string, cf map[string]map[string]string) (err error) {
	key = fmt.Sprintf(appConfigKey, brandId, platform, key)
	err = RedisStore.Set(key, cf, appConfigExpiryTime)
	if err != nil {
		return err
	}
	return nil
}
