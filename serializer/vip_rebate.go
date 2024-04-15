package serializer

import (
	"encoding/json"
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
		if idx, ok := cats[item.GameVendorBrand.CategoryId]; ok {
			if idx2, ok := catsCol[item.GameVendorBrand.CategoryId][item.GameVendorBrand.ID]; ok {
				ret.Categories[idx].Columns[idx2].Values = append(ret.Categories[idx].Columns[idx2].Values, Val)
			} else {
				Col := VipRebateColumn{
					Header:       item.GameVendorBrand.Name,
					GameVendorId: item.GameVendorId,
				}
				ret.Categories[idx].Columns = append(ret.Categories[idx].Columns, Col)
				idx2 = len(ret.Categories[idx].Columns) - 1
				ret.Categories[idx].Columns[idx2].Values = append(ret.Categories[idx].Columns[idx2].Values, Val)
				catsCol[item.GameVendorBrand.CategoryId][item.GameVendorBrand.ID] = idx2
			}
		} else {
			Cat := VipRebateCategory{
				Header:     item.GameVendorBrand.GameCategory.Name,
				CategoryId: item.GameVendorBrand.CategoryId,
			}
			ret.Categories = append(ret.Categories, Cat)
			idx = len(ret.Categories) - 1
			cats[item.GameVendorBrand.CategoryId] = idx
			catsCol[item.GameVendorBrand.CategoryId] = make(map[int64]int)
			Col := VipRebateColumn{
				Header:       item.GameVendorBrand.Name,
				GameVendorId: item.GameVendorId,
			}
			ret.Categories[idx].Columns = append(ret.Categories[idx].Columns, Col)
			idx2 := len(ret.Categories[idx].Columns) - 1
			ret.Categories[idx].Columns[idx2].Values = append(ret.Categories[idx].Columns[idx2].Values, Val)
			catsCol[item.GameVendorBrand.CategoryId][item.GameVendorBrand.ID] = len(ret.Categories[idx].Columns[idx2].Values) - 1
		}
	}
	Cat := VipRebateCategory{
		Header: "cap",
		Columns: []VipRebateColumn{
			{Header: "cap", Values: util.MapSlice(vips, func(input models.VIPRule) VipRebateColumnValue {
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
