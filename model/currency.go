package model

import (
	"gorm.io/gorm"
)

type Currency struct {
	Base
	Name			string
	DecimalPlace	int64
	FbId			int64
	DeletedAt		gorm.DeletedAt
}

func (Currency) TableName() string {
	return "currencies"
}