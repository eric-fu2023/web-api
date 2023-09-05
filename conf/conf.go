package conf

import (
	"github.com/joho/godotenv"
	"os"
	"strings"
	"web-api/cache"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/util"
)

// Init 初始化配置项
func Init() {
	os.Setenv("TZ", "Etc/GMT")

	// 从本地读取环境变量
	godotenv.Load()

	if os.Getenv("CHAT_WELCOME_NAMES") != "" {
		arr := strings.Split(os.Getenv("CHAT_WELCOME_NAMES"), "|")
		consts.ChatSystem["names"] = arr
	}
	if os.Getenv("CHAT_WELCOME_MESSAGES") != "" {
		arr := strings.Split(os.Getenv("CHAT_WELCOME_MESSAGES"), "|")
		consts.ChatSystem["messages"] = arr
	}

	// 设置日志级别
	util.BuildLogger(os.Getenv("LOG_LEVEL"))

	// 连接数据库
	replicaConn := os.Getenv("POSTGRES_DSN")
	if os.Getenv("MYSQL_REPLICA_DSN") != "" {
		replicaConn = os.Getenv("MYSQL_REPLICA_DSN")
	}
	model.Database(os.Getenv("POSTGRES_DSN"), replicaConn)
	cache.Redis()
	cache.RedisSession()
	cache.RedisShare()
	cache.RedisSyncTransaction()
	cache.SetupRedisStore()
	model.SetupMongo(os.Getenv("MONGO_URI"))

	util.InitFbFactory()
	util.InitSabaFactory()
}
