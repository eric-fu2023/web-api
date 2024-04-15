package conf

import (
	"io/ioutil"
	"log"
	"os"
	"strings"
	"web-api/util/i18n"

	yaml "gopkg.in/yaml.v2"
)

var Dictinary *map[interface{}]interface{}
var i18nDefault map[string]i18n.I18n
var defaultLocale string

func InitLocale() {
	i18nDefault = make(map[string]i18n.I18n)
	defaultLocale = os.Getenv("LANGUAGE")

	files, err := os.ReadDir("conf/locales/")
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		name := f.Name()
		lang := strings.TrimSuffix(name, ".yaml")
		i17on := i18n.I18n{}
		i17on.LoadLanguages(lang)
		i18nDefault[lang] = i17on
	}
}

func GetI18N(lang string) i18n.I18n {
	if i17on, exists := i18nDefault[lang]; exists {
		return i17on
	} else if i17on, exists := i18nDefault[defaultLocale]; exists {
		return i17on
	} else {
		return i18nDefault["en"]
	}
}

func LoadLocales(path string, dict *map[interface{}]interface{}) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	m := make(map[interface{}]interface{})
	err = yaml.Unmarshal([]byte(data), &m)
	if err != nil {
		return err
	}

	*dict = m

	return nil
}

func T(key string) string {
	dic := *Dictinary
	keys := strings.Split(key, ".")
	for index, path := range keys {
		if len(keys) == (index + 1) {
			for k, v := range dic {
				if k, ok := k.(string); ok {
					if k == path {
						if value, ok := v.(string); ok {
							return value
						}
					}
				}
			}
			return path
		}
		for k, v := range dic {
			if ks, ok := k.(string); ok {
				if ks == path {
					if dic, ok = v.(map[interface{}]interface{}); ok == false {
						return path
					}
				}
			} else {
				return ""
			}
		}
	}

	return ""
}
