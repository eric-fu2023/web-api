package model

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"gorm.io/gorm"
	"os"
	"strconv"
	"time"
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

func UserSignature(userId int64) string {
	signatureHash := md5.Sum([]byte(fmt.Sprintf("%d%s", userId, os.Getenv("USER_SIGNATURE_SALT"))))
	return hex.EncodeToString(signatureHash[:])
}
