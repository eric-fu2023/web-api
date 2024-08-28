package model

import (
	"sync"
	"time"
)

var (
	appConfigCache            = map[string]map[string]string{}
	appConfigCacheLastUpdate  int64
	appConfigCacheTimeMinutes = 10
	cacheMutex                sync.Mutex
)

func GetAppConfigWithCache(name, key string) (string, error) {
	checkIfNeedToExpireCache()

	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	if v, ok := appConfigCache[name]; ok {
		if vv, ok := v[key]; ok {
			return vv, nil
		}
	}
	var value string
	err := DB.Table("app_configs").Select("value").Where("name", name).Where("key", key).Scan(&value).Error
	if err != nil {
		return "", err
	}

	if appConfigCache[name] == nil {
		appConfigCache[name] = map[string]string{}
	}

	appConfigCache[name][key] = value
	return value, nil
}

func GetAppConfig(name, key string) (string, error) {
	var value string
	err := DB.Table("app_configs").Select("value").Where("name", name).Where("key", key).Scan(&value).Error
	if err != nil {
		return "", err
	}
	return value, nil
}

func checkIfNeedToExpireCache() {

	cacheLifeSpanSeconds := int64(appConfigCacheTimeMinutes * 60)

	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	if appConfigCacheLastUpdate == 0 || time.Now().UTC().Unix()-appConfigCacheLastUpdate > cacheLifeSpanSeconds {
		// Cache expired
		appConfigCacheLastUpdate = time.Now().UTC().Unix()

		// Empty cache
		appConfigCache = map[string]map[string]string{}
	}
}
