package model

var (
	appConfigCache = map[string]map[string]string{}
)

func GetAppConfigWithCache(name, key string) (string, error) {
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
