package serializer

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
)

type GameVendor struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	GpCode  string `json:"gp_code,omitempty"`
	WebIcon string `json:"web_icon,omitempty"`
	AppIcon string `json:"app_icon,omitempty"`
	Status  int64  `json:"status"`
}

func BuildGameVendor(a ploutos.GameVendor) (b GameVendor) {
	b = GameVendor{
		ID:      a.ID,
		Name:    a.Name,
		GpCode:  a.GameCode,
		WebIcon: Url(a.WebIcon),
		AppIcon: Url(a.AppIcon),
		Status:  a.Status,
	}
	return
}
