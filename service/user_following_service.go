package service

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"web-api/model"
	"web-api/serializer"
	"web-api/util/i18n"
)

type UserFollowingListService struct {
}

func (service *UserFollowingListService) Ids(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")
	user := u.(model.User)

	var followings []ploutos.UserFollowing
	err = model.DB.Scopes(model.UserFollowingsByUserId(user.ID)).Find(&followings).Error
	if err != nil {
		r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}
	var ids []int64
	for _, following := range followings {
		ids = append(ids, following.StreamerId)
	}

	r = serializer.Response{
		Data: ids,
	}
	return
}

func (service *UserFollowingListService) List(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")
	user := u.(model.User)

	var followings []ploutos.UserFollowing
	err = model.DB.Preload(`Streamer`, func(db *gorm.DB) *gorm.DB {
		return db.Scopes(model.StreamerWithLiveStream)
	}).Scopes(model.UserFollowingsByUserId(user.ID)).Find(&followings).Error
	if err != nil {
		r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}
	var list []serializer.Streamer
	for _, following := range followings {
		if following.Streamer == nil {
			continue
		}
		streamer := model.Streamer{
			User: *following.Streamer,
		}
		if len(following.Streamer.LiveStreams) > 0 {
			streamer.IsLive = true
		}
		list = append(list, serializer.BuildStreamer(c, streamer))
	}

	r = serializer.Response{
		Data: list,
	}
	return
}

type UserFollowingService struct {
	StreamerId int64 `form:"streamer_id" json:"streamer_id" binding:"required"`
}

func (service *UserFollowingService) Add(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")
	user := u.(model.User)

	var following ploutos.UserFollowing
	following.UserId = user.ID
	following.StreamerId = service.StreamerId
	if rows := model.DB.Scopes(model.UserFollowingsByUserIdAndStreamerId(user.ID, service.StreamerId)).First(&following).RowsAffected; rows != 0 {
		r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("already_followed"), err)
		return
	}
	err = model.DB.Save(&following).Error
	if err != nil {
		r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}
	r.Msg = "Success"
	return
}

func (service *UserFollowingService) Remove(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")
	user := u.(model.User)

	var following ploutos.UserFollowing
	err = model.DB.Scopes(model.UserFollowingsByUserIdAndStreamerId(user.ID, service.StreamerId)).Delete(&following).Error
	if err != nil {
		r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}
	r.Msg = "Success"
	return
}
