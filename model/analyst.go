package model

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
)

type Analyst struct {
	ploutos.Analyst
}

func (Analyst) List(page, limit int) (list []Analyst, err error) {
	db := DB.Scopes(Paginate(page, limit))

	err = db.
		Where("is_active", true).
		Where("deleted_at IS NULL").
		Order("created_at DESC").
		Order("id DESC").
		Find(&list).Error
	return 
}