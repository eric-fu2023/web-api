package service

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
	"web-api/model"
	"web-api/serializer"
	"web-api/util/i18n"
)

var FavouriteType = map[string]int64{
	"sport": 1,
	"game":  2,
}

type UserFavouriteListService struct {
	Type    int64 `form:"type" json:"type" binding:"required"`
	SportId int64 `form:"sport_id" json:"sport_id"`
}

func (service *UserFavouriteListService) List(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")
	user := u.(model.User)

	var favourites []ploutos.UserFavourite
	q := model.DB.Scopes(model.UserFavouriteByUserIdTypeAndSportId(user.ID, service.Type, service.SportId))
	if service.Type == FavouriteType["game"] {
		q.Preload(`SubGameBrand`)
	}
	err = q.Find(&favourites).Order(`id DESC`).Error
	if err != nil {
		r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}
	var list []serializer.UserFavourite
	for _, favourite := range favourites {
		list = append(list, serializer.BuildUserFavourite(favourite))
	}

	r = serializer.Response{
		Data: list,
	}
	return
}

type UserFavouriteService struct {
	GameId  int64 `form:"game_id" json:"game_id" binding:"required"`
	Type    int64 `form:"type" json:"type" binding:"required"`
	SportId int64 `form:"sport_id" json:"sport_id"`
}

func (service *UserFavouriteService) Add(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")
	user := u.(model.User)

	var favourite ploutos.UserFavourite
	favourite.UserId = user.ID
	favourite.GameId = service.GameId
	favourite.Type = service.Type
	favourite.SportId = service.SportId
	if rows := model.DB.Scopes(model.UserFavouriteByUserIdTypeGameIdAndSportId(user.ID, service.Type, service.GameId, service.SportId)).First(&favourite).RowsAffected; rows != 0 {
		r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("favourite_already_added"), err)
		return
	}
	err = model.DB.Save(&favourite).Error
	if err != nil {
		r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}
	r.Msg = "Success"
	return
}

func (service *UserFavouriteService) Remove(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")
	user := u.(model.User)

	var favourite ploutos.UserFavourite
	err = model.DB.Scopes(model.UserFavouriteByUserIdTypeGameIdAndSportId(user.ID, service.Type, service.GameId, service.SportId)).Delete(&favourite).Error
	if err != nil {
		r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}
	r.Msg = "Success"
	return
}
