package model

import (
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"
)

type Base struct {
	ID        int64 `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type BrandAgent struct {
	BrandId int64
	AgentId int64
}

type ShardTable interface {
	TableName() string
	DefineShard() []int
}

func Sharded(model ShardTable, addAs bool) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		str := model.TableName()
		ids := model.DefineShard()
		for _, id := range ids {
			str += "_" + strconv.Itoa(id)
		}
		if addAs {
			str += " AS " + model.TableName()
		}
		return db.Table(str)
	}
}

func Paginate(page int, limit int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Limit(limit).Offset(limit * (page - 1))
	}
}

func Ongoing(now time.Time, startField, endField string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(fmt.Sprintf("%s < ? and %s > ?", startField, endField), now, now)
	}
}
