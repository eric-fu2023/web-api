package serializer

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"web-api/util"

	models "blgit.rfdev.tech/taya/ploutos-object"
)

type VipRebateColumnValue struct {
	VipLevel int64   `json:"vip_level"`
	Format   string  `json:"format"` //percentage,currency
	Value    float64 `json:"value"`
}

type VipRebateCategory struct {
	CategoryId int64             `json:"category_id"`
	Header     string            `json:"header"`
	Columns    []VipRebateColumn `json:"columns"`
}

type VipRebateColumn struct {
	Header       string                 `json:"header"`
	GameVendorId int64                  `json:"game_vendor_id"`
	Values       []VipRebateColumnValue `json:"values"`
}

type VipRebateDetails struct {
	Categories  []VipRebateCategory `json:"categories"`
	Description []map[string]any    `json:"description"`
}

func BuildVipRebateDetails(list []models.VipRebateRule, desc string, vips []models.VIPRule) (ret VipRebateDetails) {
	bytes := []byte(desc)
	_ = json.Unmarshal(bytes, &ret.Description)
	// get category column
	cats := map[int64]int{}
	catsCol := map[int64]map[int64]int{}
	for _, item := range list {
		Val := VipRebateColumnValue{
			VipLevel: item.VipLevel,
			Format:   "percentage",
			Value:    item.RebateRate,
		}
		if idx, ok := cats[item.GameVendor.CategoryId]; ok {
			if idx2, ok := catsCol[item.GameVendor.CategoryId][item.GameVendor.ID]; ok {
				ret.Categories[idx].Columns[idx2].Values = append(ret.Categories[idx].Columns[idx2].Values, Val)
			} else {
				Col := VipRebateColumn{
					Header:       item.GameVendor.Name,
					GameVendorId: item.GameVendorId,
				}
				ret.Categories[idx].Columns = append(ret.Categories[idx].Columns, Col)
				idx2 = len(ret.Categories[idx].Columns) - 1
				ret.Categories[idx].Columns[idx2].Values = append(ret.Categories[idx].Columns[idx2].Values, Val)
				catsCol[item.GameVendor.CategoryId][item.GameVendor.ID] = idx2
			}
		} else {
			Cat := VipRebateCategory{
				Header:     item.GameVendor.GameCategory.Name,
				CategoryId: item.GameVendor.CategoryId,
			}
			ret.Categories = append(ret.Categories, Cat)
			idx = len(ret.Categories) - 1
			cats[item.GameVendor.CategoryId] = idx
			catsCol[item.GameVendor.CategoryId] = make(map[int64]int)
			Col := VipRebateColumn{
				Header:       item.GameVendor.Name,
				GameVendorId: item.GameVendorId,
			}
			ret.Categories[idx].Columns = append(ret.Categories[idx].Columns, Col)
			idx2 := len(ret.Categories[idx].Columns) - 1
			ret.Categories[idx].Columns[idx2].Values = append(ret.Categories[idx].Columns[idx2].Values, Val)
			catsCol[item.GameVendor.CategoryId][item.GameVendor.ID] = len(ret.Categories[idx].Columns[idx2].Values) - 1
		}
	}
	for idx := range ret.Categories {
		col := VipRebateColumn{
			Header: "primary",
		}

		// based on actual data
		vipLvlMap := make(map[int64]int)
		for i := range ret.Categories[idx].Columns {
			for j := range ret.Categories[idx].Columns[i].Values {
				if vipIdx, exists := vipLvlMap[ret.Categories[idx].Columns[i].Values[j].VipLevel]; exists {
					col.Values[vipIdx].Value = max(col.Values[vipIdx].Value, ret.Categories[idx].Columns[i].Values[j].Value)
				} else {
					col.Values = append(col.Values, ret.Categories[idx].Columns[i].Values[j])
					vipLvlMap[ret.Categories[idx].Columns[i].Values[j].VipLevel] = len(col.Values) - 1
				}
			}
		}

		// // based on vip setting
		// col.Values = util.MapSlice(vips, func(input models.VIPRule) VipRebateColumnValue {
		// 	return VipRebateColumnValue{
		// 		VipLevel: input.VIPLevel,
		// 		Format:   "percentage",
		// 		Value:    input.RebateRate,
		// 	}
		// })
		ret.Categories[idx].Columns = append(ret.Categories[idx].Columns, col)
	}

	// add cap
	Cat := VipRebateCategory{
		Header: "cap",
		Columns: []VipRebateColumn{
			{Header: "primary", Values: util.MapSlice(vips, func(input models.VIPRule) VipRebateColumnValue {
				return VipRebateColumnValue{
					VipLevel: input.VIPLevel,
					Format:   "currency",
					Value:    float64(input.RebateCap / 100),
				}
			})},
		},
	}
	ret.Categories = append(ret.Categories, Cat)
	return ret
}

// BuildVipReferralDetails builds the vip referral details response
// list is sorted by game_category_id ASC, vip_level ASC
func BuildVipReferralDetails(c *gin.Context, list []models.VipReferralAllianceRule, desc string, vips []models.VIPRule) (ret VipRebateDetails) {
	bytes := []byte(desc)
	_ = json.Unmarshal(bytes, &ret.Description)
	var currentGameCategoryId int64 = 0
	for _, item := range list {
		// Init new column
		if item.GameCategoryId != currentGameCategoryId {
			ret.Categories = append(ret.Categories, VipRebateCategory{
				Header:     FormatGameCategoryName(c, item.GameCategoryId),
				CategoryId: item.GameCategoryId,
				Columns: []VipRebateColumn{{
					Header: "primary",
				}},
			})
			currentGameCategoryId = item.GameCategoryId
		}
		// Add row
		catIdx := len(ret.Categories) - 1
		ret.Categories[catIdx].Columns[0].Values = append(ret.Categories[catIdx].Columns[0].Values, VipRebateColumnValue{
			VipLevel: item.VipLevel,
			Format:   "percentage",
			Value:    item.RebateRate,
		})
	}

	// add cap
	Cat := VipRebateCategory{
		Header: "cap",
		Columns: []VipRebateColumn{
			{Header: "primary", Values: util.MapSlice(vips, func(input models.VIPRule) VipRebateColumnValue {
				return VipRebateColumnValue{
					VipLevel: input.VIPLevel,
					Format:   "currency",
					Value:    float64(input.RebateCap / 100),
				}
			})},
		},
	}
	ret.Categories = append(ret.Categories, Cat)
	return ret
}
