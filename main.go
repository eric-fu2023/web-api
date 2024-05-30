package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"web-api/conf"
	"web-api/model"
	"web-api/server"
	"web-api/task"
	websocketTask "web-api/task/websocket"
	"web-api/util"
	"web-api/websocket"

	"github.com/gin-contrib/pprof"
	"github.com/robfig/cron/v3"
)

var runTask bool

func init() {
	runTask = false
	argsWithoutProg := os.Args
	command := ""
	if len(os.Args) > 1 {
		command = argsWithoutProg[1]
		if command != "run-task" {
			log.Fatalf("command:" + command + " not found, do you mean run-task?")
		} else {
			runTask = true
		}
	}
}

func main() {
	conf.Init()

	if runTask {
		go task.ProcessFbSyncTransaction()
		go task.ProcessTayaSyncTransaction()
		go task.ProcessSabaSettle()
		go task.ProcessImUpdateBalance()
		go task.ConsumeMgStreams()
		go task.ConsumeMgStreamsHot()
		if os.Getenv("MQTT_ADDRESS") != "" { // mqtt tasks
			go task.UpdateOnlineStatus()
			go task.UpdateUnsubscribed()
			go task.UpdateSubscribed()
			util.InitMQTT()
		}
		go func() {
			websocketTask.Functions = []func(*websocket.Connection, context.Context, context.CancelFunc){ // modules to be run when connected
				websocketTask.Reply,
				websocketTask.Event,
			}
			websocketTask.Connect(10)
		}()
		c := cron.New(cron.WithSeconds())
		c.AddFunc("0 */5 * * * *", func() {
			task.RefreshPaymentOrder()
		})
		//c.AddFunc("10 */1 * * * *", func() {
		//	task.CalculateSortFactor()
		//})
		c.Start()
		select {}
	} else {
		//task.CreateUserWallet([]int64{8, 9}, 1) // to create wallets when a new game vendor is added
		//task.CreateImOneUsersForExistingTayaUsers() // to create wallets when a new game vendor is added // to create wallets when a new game vendor is added
		//task.EncryptMobileAndEmail()
		task.SetRandomAvatar()
		if os.Getenv("MQTT_ADDRESS") != "" {
			util.InitMQTT()
		}
		go task.SendDummyMqtt() // to avoid mqtt from disconnecting
		r := server.NewRouter()
		pprof.Register(r)
		srv := &http.Server{
			Addr:    ":" + os.Getenv("PORT"),
			Handler: r,
		}
		go func() {
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("listen: %s\n", err)
			}
		}()
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Fatal("Server forced to shutdown: ", err)
		}
		model.GlobalWaitGroup.Wait()
		log.Println("Server gracefully shutdown")
	}

	defer model.IPDB.Close()
}
