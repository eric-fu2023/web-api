package model

import (
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

// DB 数据库链接单例
var DB *gorm.DB
var MongoDB *mongo.Database
var IPDB *awdb.Reader

// Database 在中间件中初始化postgres链接
func Database(primaryConn string, replicaConn string) {
	// 初始化GORM日志配置
	getLogLevel := func () logger.LogLevel {
		if os.Getenv("ENV") == "local" {
			return logger.Info
		}
		return logger.Warn
	}
	newLogger := NewCustomLogger(os.Stdout,logger.Config{
		SlowThreshold:             2*time.Second, // Slow SQL threshold
		LogLevel:                  getLogLevel(), // Log level(这里记得根据需求改一下)
		IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
		Colorful:                  false,       // Disable color
	})

	db, err := gorm.Open(postgres.Open(primaryConn), &gorm.Config{
		Logger: newLogger,
	})
	db.Use(dbresolver.Register(dbresolver.Config{ // for other models and tables, use primary db for both reads and writes
		Sources:  []gorm.Dialector{postgres.Open(primaryConn)},
		Replicas: []gorm.Dialector{postgres.Open(primaryConn)},
	}).Register(dbresolver.Config{ // for `matches` and `news`, use replica for reads and primary for writes
		Sources:  []gorm.Dialector{postgres.Open(primaryConn)},
		Replicas: []gorm.Dialector{postgres.Open(replicaConn)},
	}).
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
