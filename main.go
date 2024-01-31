package main

import (
	"context"
	"github.com/gin-contrib/pprof"
	"github.com/robfig/cron/v3"
	"log"
	"os"
	"web-api/conf"
	"web-api/model"
	"web-api/server"
	"web-api/task"
	websocketTask "web-api/task/websocket"
	"web-api/websocket"
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
		go func() {
			websocketTask.Functions = []func(*websocket.Connection, context.Context, context.CancelFunc){ // modules to be run when connected
				websocketTask.Reply,
			}
			websocketTask.Connect(10)
		}()

		c := cron.New(cron.WithSeconds())
		c.AddFunc("0 */5 * * * *", func() {
			task.RefreshPaymentOrder()
		})
		c.AddFunc("10 */1 * * * *", func() {
			task.CalculateSortFactor()
		})
		c.Start()
		select {}
	} else {
		r := server.NewRouter()
		pprof.Register(r)
		r.Run(":" + os.Getenv("PORT"))
	}

	defer model.IPDB.Close()
}
