package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"os"
	"strings"
	"web-api/util/i18n"
)

func Locale() gin.HandlerFunc {
	return func(c *gin.Context) {
		locale := "en-us"
		language := "en"
		countryCode := "us"
		if os.Getenv("PLATFORM_LANGUAGE") != "" {
			locale = os.Getenv("PLATFORM_LANGUAGE")
			language = locale
		}
		if c.GetHeader("Locale") != "" {
			l := strings.ToLower(c.GetHeader("Locale"))
			locale = l
			ll := strings.Split(l, "-")
			if len(ll) > 0 {
				language = ll[0]
			}
			if len(ll) > 1 {
				countryCode = ll[1]
			}
		}
		i18n := i18n.I18n{}
		if err := i18n.LoadLanguages(language); err != nil {
			fmt.Println(err)
		}
		c.Set("i18n", i18n)
		c.Set("_locale", locale)
		c.Set("_language", language)
		c.Set("_country_code", countryCode)
		c.Next()
	}
}
