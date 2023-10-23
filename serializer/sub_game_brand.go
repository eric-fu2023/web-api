package serializer

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"time"
)

type SubGameBrand struct {
	Id               int64            `json:"id"`
	Name             string           `json:"name"`
	SubGameId        int64            `json:"sub_game_id"`
	GameCode         string           `json:"game_code"`
	GameType         string           `json:"game_type"`
	WebIcon          string           `json:"web_icon,omitempty"`
	AppIcon          string           `json:"app_icon,omitempty"`
	IsMaintenance    bool             `json:"is_maintenance,omitempty"`
	MaintenanceStart int64            `json:"maintenance_start,omitempty"`
	MaintenanceEnd   int64            `json:"maintenance_end,omitempty"`
	Vendor           *GameVendorBrand `json:"vendor,omitempty"`
}

func BuildSubGameBrand(a ploutos.SubGameBrand) (b SubGameBrand) {
	b = SubGameBrand{
		Id:        a.ID,
		Name:      a.Name,
		SubGameId: a.SubGameId,
		GameCode:  a.GameCode,
		GameType:  a.GameType,
		WebIcon:   Url(a.WebIcon),
		AppIcon:   Url(a.AppIcon),
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
	if a.GameVendorBrand != nil {
		t := BuildGameVendorBrand(*a.GameVendorBrand)
		b.Vendor = &t
	}
	return
}
