package service

import (
	"github.com/gin-gonic/gin"
	"web-api/model"
	"web-api/serializer"
	"web-api/util/i18n"
)

type CategoryListService struct {
}

func (service *CategoryListService) List(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	var categories []model.CategoryType
	if err = model.DB.Model(model.CategoryType{}).Scopes(model.CategoryTypeWithCategories).Find(&categories).Error; err != nil {
		r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}
	var list []serializer.CategoryType
	for _, category := range categories {
		list = append(list, serializer.BuildCategoryType(category))
	}

	r = serializer.Response{
		Data: list,
	}
	return
}
