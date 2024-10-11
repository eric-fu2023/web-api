package api

import (
	"web-api/model"
	"web-api/serializer"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"github.com/gin-gonic/gin"
)

type HomeBannerService struct{}

func GetHomeBanners(c *gin.Context) {
	var service HomeBannerService
	if err := c.ShouldBind(&service); err == nil {
		res := service.Get(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

type HomeBanner struct {
	Id             int64  `json:"id" gorm:"column:id"`
	ImgUrl         string `json:"img_url" gorm:"column:img_url"`
	NavigationType string `json:"navi_type" gorm:"column:navi_type"`
	NavigationId   int64  `json:"navi_id" gorm:"column:navi_id"`
	Iframe         bool   `json:"iframe" gorm:"column:iframe"`
	LoginRequired  bool   `json:"login_required" gorm:"column:login_required"`
}

type HomeBannerServiceGetResponse struct {
	Banners []HomeBanner `json:"banners"`
}

func (service *HomeBannerService) Get(c *gin.Context) serializer.Response {
	// ip := retrieveClientIp(c)
	var banners []ploutos.HomeBanner
	err := model.DB.Find(&banners).Error
	if err != nil {
		r := serializer.Err(c, service, serializer.CodeGeneralError, "error get home banners", err)
		return r
	}

	var bannersR []HomeBanner
	for _, banner := range banners {
		bannersR = append(bannersR, HomeBanner{
			Id:             banner.ID,
			ImgUrl:         banner.ImgUrl,
			NavigationType: banner.NavigationType,
			NavigationId:   banner.NavigationId,
			Iframe:         banner.Iframe,
			LoginRequired:  banner.LoginRequired,
		})
	}

	return serializer.Response{
		Msg: "success",
		Data: HomeBannerServiceGetResponse{
			Banners: bannersR,
		},
	}
}
