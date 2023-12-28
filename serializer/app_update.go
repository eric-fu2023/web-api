package serializer

import (
	"web-api/model"
)

type AppUpdate struct {
	Url           string `json:"url"`
	IsForce       bool   `json:"is_force"`
	Version       string `json:"version"`
	VersionSerial int64  `json:"version_serial"`
	Remark        string `json:"remark,omitempty"`
}

func BuildAppUpdate(a model.AppUpdate) (b AppUpdate) {
	b = AppUpdate{
		Url:           a.Url,
		IsForce:       a.IsForce,
		Version:       a.Version,
		VersionSerial: a.VersionSerial,
		Remark:        a.Remark,
	}
	return
}
