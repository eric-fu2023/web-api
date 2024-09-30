package game_integration

import (
	"fmt"
	"slices"
	"time"

	"web-api/conf/consts"
	"web-api/model"
	"web-api/serializer"
	"web-api/util/i18n"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"github.com/gin-gonic/gin"
)

type SubGameService struct {
	Platform int64 `form:"platform" json:"platform" binding:"required"`
}

// gameTypeOrdering for (service *SubGameService) List
var gameTypeOrdering = map[string]int{
	"LIVE":   0,
	"CRASH":  1,
	"FLASH":  2,
	"SPRIBE": 3,
	"BOARD":  4,
	"SLOTS":  5,
	"TABLE":  6,
}

// SubGameBrand for (service *SubGameService) List
type SubGameBrand struct {
	ploutos.BASE
	Name          string    `json:"name" form:"name" gorm:"column:name"`
	VendorBrandId int64     `json:"vendor_brand_id" form:"vendor_brand_id" gorm:"column:vendor_brand_id"`
	WebIcon       string    `json:"web_icon" form:"web_icon" gorm:"column:web_icon"`
	AppIcon       string    `json:"app_icon" form:"app_icon" gorm:"column:app_icon"`
	Web           int64     `json:"web" form:"web" gorm:"column:web"`
	H5            int64     `json:"h5" form:"h5" gorm:"column:h5"`
	Ios           int64     `json:"ios" form:"ios" gorm:"column:ios"`
	Android       int64     `json:"android" form:"android" gorm:"column:android"`
	Status        int64     `json:"status" form:"status" gorm:"column:status"`
	BrandId       int64     `json:"brand_id" form:"brand_id" gorm:"column:brand_id"`
	SubGameId     int64     `json:"sub_game_id" form:"sub_game_id" gorm:"column:sub_game_id"`
	GameCode      string    `json:"game_code" form:"game_code" gorm:"column:game_code"`
	GameType      string    `json:"game_type" form:"game_type" gorm:"column:game_type"`
	IsFeatured    bool      `json:"is_featured" form:"is_featured" gorm:"column:is_featured"`
	StartTime     time.Time `json:"start_time" form:"start_time" gorm:"column:start_time"`
	EndTime       time.Time `json:"end_time" form:"end_time" gorm:"column:end_time"`
	SortRanking   int64     `json:"sort" form:"sort" gorm:"column:sort"`

	GameVendorBrand *ploutos.GameVendorBrand `json:"game_vendor_brand,omitempty" form:"-" gorm:"references:VendorBrandId;foreignKey:ID"`

	SubGame *ploutos.SubGameC `json:"game_vendor_brand,omitempty" form:"-" gorm:"references:SubGameId;foreignKey:ID"`
}

func (SubGameBrand) TableName() string {
	return ploutos.TableNameSubGameBrand
}

func (service *SubGameService) List(c *gin.Context) (serializer.Response, error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	brandId := c.MustGet("_brand").(int)

	platform, ok := consts.PlatformIdToGameVendorColumn[service.Platform]
	if !ok {
		r := serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("invalid_platform"), nil)
		return r, nil
	}

	var brandSubGames []SubGameBrand
	tx := model.DB.Model(SubGameBrand{}).Preload("GameVendorBrand").
		Joins(fmt.Sprintf(`LEFT JOIN game_vendor_brand gvb on gvb.id = %s.vendor_brand_id`, SubGameBrand{}.TableName())).
		Preload("SubGame").
		Joins(fmt.Sprintf(`LEFT JOIN sub_game sg on sg.id = %s.sub_game_id`, SubGameBrand{}.TableName())).
		Where(fmt.Sprintf("%s.brand_id = %d", SubGameBrand{}.TableName(), brandId)).
		Where(fmt.Sprintf("%s.%s = %d", SubGameBrand{}.TableName(), platform, 1)).Find(&brandSubGames)
	if err := tx.Error; err != nil {
		return serializer.Response{
			Data: []serializer.SubGamesBrandsByGameType{},
		}, err
	}

	// fixme this is all items sort, can combine with group sorting in BuildBrandSubGamesByGameType
	slices.SortFunc(brandSubGames, func(a, b SubGameBrand) int {
		return int(a.SortRanking - b.SortRanking)
	})

	buildForm := make([]serializer.SubGameBrandRaw, 0, len(brandSubGames))
	for _, bsg := range brandSubGames {
		buildForm = append(buildForm, serializer.SubGameBrandRaw(bsg))
	}

	data, err := serializer.BuildBrandSubGamesByGameType(buildForm, gameTypeOrdering)
	if err != nil {
		return serializer.Response{
			Data: []serializer.SubGamesBrandsByGameType{},
		}, err
	}

	return serializer.Response{
		Data: data,
	}, nil
}

func (service *SubGameService) FeaturedList(c *gin.Context) (serializer.Response, error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	brandId := c.MustGet("_brand").(int)

	platform, ok := consts.PlatformIdToGameVendorColumn[service.Platform]
	if !ok {
		r := serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("invalid_platform"), nil)
		return r, nil
	}

	// todo use view model from this repo
	var subGames []ploutos.SubGameBrand
	tx := model.DB.Debug().Model(ploutos.SubGameBrand{}).Where(fmt.Sprintf("%s.brand_id = %d", ploutos.SubGameBrand{}.TableName(), brandId)).Where(fmt.Sprintf("%s.%s = ?", ploutos.SubGameBrand{}.TableName(), platform), 1).Where("is_featured = true").Order("updated_at").Find(&subGames)
	if err := tx.Error; err != nil {
		return serializer.Response{
			Data: []serializer.SubGamesBrandsByGameType{},
		}, err
	}

	data := serializer.BuildFeaturedGames(subGames)

	return serializer.Response{
		Data: data,
	}, nil
}
