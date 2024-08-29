package i18n

import (
	"os"
	"strings"

	"golang.org/x/text/language"
	"golang.org/x/text/message"

	yaml "gopkg.in/yaml.v2"
)

type I18n struct {
	Language   string
	Dictionary *map[interface{}]interface{}
}

func (i18n *I18n) LoadLanguages(locale string) error {
	data, err := os.ReadFile("conf/locales/" + locale + ".yaml")
	if err != nil {
		return err
	}

	m := make(map[interface{}]interface{})
	err = yaml.Unmarshal([]byte(data), &m)
	if err != nil {
		return err
	}

	i18n.Language = locale
	i18n.Dictionary = &m

	return nil
}

func (i18n *I18n) T(key string) string {
	dic := *i18n.Dictionary
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

func (i18n *I18n) FormatCurrencyAndValue(i int64) (r string) {
	p := message.NewPrinter(language.English)
	if i18n.Language == "zh" {
		if i < 10000 {
			r = p.Sprintf("%d欧", i)
		} else if i < 100000000 {
			r = p.Sprintf("%d万欧", i/10000)
		} else {
			r = p.Sprintf("%d亿 %d万欧", i/100000000, (i-(i/100000000)*100000000)/10000)
		}
	} else {
		if i < 1000 {
			r = p.Sprintf("%d€", i)
		} else if i < 1000000 {
			r = p.Sprintf("%dK€", i/1000)
		} else {
			r = p.Sprintf("%dM€", i/1000000)
		}
	}
	return
}
