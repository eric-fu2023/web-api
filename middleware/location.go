package middleware

import (
	"github.com/gin-gonic/gin"
	"web-api/model"
	"net"
)

func Location() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := net.ParseIP(c.ClientIP())
		var record map[string]interface{}
		err := model.IPDB.Lookup(ip, &record)
		if err != nil {
			//fmt.Println(err)
		}
		if len(record) == 0 { // if ip is not found in country/city db, no prefix
			return
		}

		country := string(record["country"].([]byte))
		province := string(record["province"].([]byte))
		city := string(record["city"].([]byte))
		c.Set("_country", country)
		c.Set("_province", province)
		c.Set("_city", city)
		c.Next()
	}
}