package service

import (
	"web-api/model"
	"web-api/serializer"
	"web-api/util/i18n"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
)

type CategoryListService struct {
}

func (service *CategoryListService) List(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	var categories []ploutos.CategoryType
	if err = model.DB.Model(ploutos.CategoryType{}).Scopes(model.CategoryTypeWithCategories).Find(&categories).Error; err != nil {
		r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}
	var list []serializer.CategoryType
	for _, category := range categories {
		list = append(list, serializer.BuildCategoryType(category))
	}

	isA := false
	if v, exists := c.Get("_isA"); exists {
		if vv, ok := v.(bool); ok && vv {
			isA = vv
		}
	}
	if isA {
		filteredList := []serializer.CategoryType{} // Assuming list is of type []serializer.ListItem

		for _, listItem := range list {
			if listItem.ID == 3 { // Condition for "batace"
				filteredCategories := []serializer.Category{}

				for _, category := range listItem.Categories {
					if category.ID == 4 { // Condition for "cricket"
						filteredCategories = append(filteredCategories, category)
						break // Stop after finding the matching category
					}
				}

				// If the listItem has the required category, update and keep it in the list
				if len(filteredCategories) > 0 {
					listItem.Categories = filteredCategories
					filteredList = append(filteredList, listItem)
				}

			}

		}

		// Replace the original list with the filtered version
		list = filteredList
	}

	r = serializer.Response{
		Data: list,
	}
	return
}
