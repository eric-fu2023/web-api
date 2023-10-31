package service

import (
	"github.com/gin-gonic/gin"
	"web-api/model"
	"web-api/serializer"
)

type CounterService struct {
}

func (service *CounterService) Get(c *gin.Context) serializer.Response {
	//u, _ := c.Get("user")
	//user := u.(model.User)
	counters := model.UserCounters{
		Order:        100,
		Transaction:  0,
		Notification: 11,
	}
	return serializer.Response{
		Data: serializer.BuildUserCounters(c, counters),
	}
}
