package serializer

import (
	"time"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
)

type GameVendorBrand struct {
	ID               int64  `json:"id"`
	Name             string `json:"name"`
	Title            string `json:"title"`
	GameId           int64  `json:"game_id,omitempty"`
	Type             int64  `json:"type"`
	WebIcon          string `json:"web_icon,omitempty"`
	AppIcon          string `json:"app_icon,omitempty"`
	SmallIcon        string `json:"small_icon"`
	IsMaintenance    bool   `json:"is_maintenance,omitempty"`
	MaintenanceStart int64  `json:"maintenance_start,omitempty"`
	MaintenanceEnd   int64  `json:"maintenance_end,omitempty"`
}

func BuildGameVendorBrand(a ploutos.GameVendorBrand) (b GameVendorBrand) {
	b = GameVendorBrand{
		ID:        a.GameVendorId,
		Name:      a.Name,
		Title:     a.Title,
		Type:      a.CategoryId,
		WebIcon:   Url(a.WebIcon),
		AppIcon:   Url(a.AppIcon),
		SmallIcon: Url(a.SmallIcon),
	}
	if !a.StartTime.IsZero() {
		if time.Now().After(a.StartTime) && (time.Now().Before(a.EndTime) || a.EndTime.IsZero()) {
			b.IsMaintenance = true
			b.MaintenanceStart = a.StartTime.Unix()
			if !a.EndTime.IsZero() {
				b.MaintenanceEnd = a.EndTime.Unix()
			}
		}
	}
	return
}

// only used for GameCategoryListService
func BuildGameVendorBrandForCategory(a ploutos.GameVendorBrand, gameId int64) (b GameVendorBrand) {
	b = GameVendorBrand{
		ID:        a.GameVendorId,
		Name:      a.Name,
		Title:     a.Title,
		GameId:    gameId,
		Type:      a.CategoryId,
		WebIcon:   Url(a.WebIcon),
		AppIcon:   Url(a.AppIcon),
		SmallIcon: Url(a.SmallIcon),
	}
	if !a.StartTime.IsZero() {
		if time.Now().After(a.StartTime) && (time.Now().Before(a.EndTime) || a.EndTime.IsZero()) {
			b.IsMaintenance = true
			b.MaintenanceStart = a.StartTime.Unix()
			if !a.EndTime.IsZero() {
				b.MaintenanceEnd = a.EndTime.Unix()
			}
		}
	}
	return
}
