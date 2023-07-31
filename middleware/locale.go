package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"strings"
	"web-api/util/i18n"
)

func Locale() gin.HandlerFunc {
	return func(c *gin.Context) {
		locale := "en-us"
		language := "en"
		if c.GetHeader("Locale") != "" {
			l := strings.ToLower(c.GetHeader("Locale"))
			locale = l
			ll := strings.Split(l, "-")
			language = ll[0]
		}
		i18n := i18n.I18n{}
		if err := i18n.LoadLanguages(language); err != nil {
			fmt.Println(err)
		}
		c.Set("i18n", i18n)
		c.Set("_locale", locale)
		c.Set("_language", language)
		c.Next()
	}
}
