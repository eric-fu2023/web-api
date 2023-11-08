package model

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/driver/postgres"
	"gorm.io/plugin/dbresolver"
	"os"
	"time"
	"web-api/util"
	"web-api/util/awdb"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var txRelated = []any{&ploutos.UserSum{}, &ploutos.UserSumC{}, "user_sums",
	&ploutos.Transaction{}, &ploutos.DcTransaction{}, &ploutos.FbTransaction{}, &ploutos.SabaTransaction{},
	&ploutos.TransactionC{}, &ploutos.DcTransactionC{}, &ploutos.FbTransactionC{}, &ploutos.SabaTransactionC{},
	&ploutos.CashOrder{}, &ploutos.CashOrderC{}, &CashOrder{}, "cash_orders",
	"transactions", "fb_transactions", "dc_transactions", "saba_transactions", "txConn",
}

var DB *gorm.DB
var MongoDB *mongo.Database
var IPDB *awdb.Reader

func Database(primaryConn string, txConn string) {
	getLogLevel := func() logger.LogLevel {
		if os.Getenv("ENV") == "local" {
			return logger.Info
		}
		return logger.Warn
	}
	newLogger := NewCustomLogger(os.Stdout, logger.Config{
		SlowThreshold:             2 * time.Second, // Slow SQL threshold
		LogLevel:                  getLogLevel(),   // Log level
		IgnoreRecordNotFoundError: true,            // Ignore ErrRecordNotFound error for logger
		Colorful:                  false,           // Disable color
	})

	db, err := gorm.Open(postgres.Open(primaryConn), &gorm.Config{
		Logger: newLogger,
	})
	db.Use(dbresolver.Register(dbresolver.Config{
		Sources: []gorm.Dialector{postgres.Open(txConn)},
	}, txRelated...).
		SetMaxOpenConns(75).
		SetMaxIdleConns(25).
		SetConnMaxLifetime(time.Hour))

	// Error
	if primaryConn == "" || err != nil {
		util.Log().Error("postgres lost: %v", err)
		panic(err)
	}
	_, err = db.DB()
	if err != nil {
		util.Log().Error("postgres lost: %v", err)
		panic(err)
	}

	DB = db

	//migration()

	IPDB, err = awdb.Open("./IP_city_single_WGS84.awdb")
	if err != nil {
		panic(err)
	}
}

func SetupMongo(uri string) {
	ctx := context.TODO()
	if client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri)); err != nil {
		panic(err)
	} else {
		MongoDB = client.Database(os.Getenv("MONGO_DB"))
	}
}
