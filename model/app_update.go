package model

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"strconv"
	"strings"
)

type AppUpdate struct {
	ploutos.AppUpdate
}

func (a *AppUpdate) Get(brandId int64, osType int64, version string, channel string) (err error) {
	query := DB.Model(&a).
		Where(`status = 1`).
		Where(`brand_id`, brandId).
		Where(`os_type`, osType).
		Where(`channel`, channel).
		Order(`version_serial DESC`)

	ver := strings.Split(version, ".")
	var newVer string
	for i := range []int{0, 1, 2} {
		var n int
		n, err = strconv.Atoi(ver[i])
		if err != nil {
			return
		}
		if n < 10 {
			newVer += "0"
		}
		newVer += ver[i]
	}
	var n int
	n, err = strconv.Atoi(newVer)
	if err != nil {
		return
	}
	query = query.Where(`version_serial > ?`, n)

	err = query.Limit(1).Find(a).Error

	return
}
