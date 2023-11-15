package service

import (
	"github.com/gin-gonic/gin"
	"web-api/model"
	"web-api/serializer"
	"web-api/service/common"
)

type RoomChatHistoryService struct {
	RoomId   string `form:"room_id" json:"room_id" binding:"required"`
	TimeFrom int64  `form:"time_from" json:"time_from"`
	common.Page
}

func (service *RoomChatHistoryService) List(c *gin.Context) (r serializer.Response, err error) {
	var m model.RoomMessage
	messages, err := m.List(service.RoomId, service.TimeFrom, int64(service.Page.Page), int64(service.Limit))
	var list []serializer.RoomMessage
	for _, msg := range messages {
		list = append(list, serializer.BuildRoomMessage(msg))
	}
	r = serializer.Response{
		Data: list,
	}
	return
}
