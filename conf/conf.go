package conf

import (
	"os"
	"strings"
	"web-api/cache"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/service/aj_captcha"
	"web-api/util"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-gorm/caches/v3"
	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
)

func Init() {
	os.Setenv("TZ", "Etc/GMT")

	godotenv.Load()
	InitCfg()
	InitLocale()

	if os.Getenv("CHAT_WELCOME_NAMES") != "" {
		arr := strings.Split(os.Getenv("CHAT_WELCOME_NAMES"), "|")
		consts.ChatSystem["names"] = arr
	}
	if os.Getenv("CHAT_WELCOME_MESSAGES") != "" {
		arr := strings.Split(os.Getenv("CHAT_WELCOME_MESSAGES"), "|")
		consts.ChatSystem["messages"] = arr
	}

	util.BuildLogger(os.Getenv("LOG_LEVEL"))

	model.Database(os.Getenv("POSTGRES_DSN"), os.Getenv("POSTGRES_TX_DSN"))
	cache.Redis()
	cache.RedisSession()
	cache.RedisShare()
	cache.RedisSyncTransaction()
	cache.RedisConfig()
	cache.SetupRedisStore()
	cache.RedisLock()
	cache.RedisRecentGames()
	model.SetupMongo(os.Getenv("MONGO_URI"))
	aj_captcha.Init()

	util.InitTayaFactory()
	util.InitFbFactory()
	util.InitSabaFactory()
	util.InitDcFactory()
	util.InitImFactory()
	util.InitUgsFactory()
	util.InitImOneFactory()

	util.InitFCMFactory()
	model.InitShengWang()

	if os.Getenv("MQTT_ADDRESS") != "" {
		util.InitMQTT()
	}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("username", util.UsernameValidation)
		v.RegisterValidation("password", util.PasswordValidation)
	}
	loadGormCaches(cache.RedisClient)
}

func loadGormCaches(client *redis.Client) {
	cachesPlugin := &caches.Caches{Conf: &caches.Config{
		Cacher: &model.RedisCacher{
			Redis: client,
		},
	}}
	model.DB.Use(cachesPlugin)
}
