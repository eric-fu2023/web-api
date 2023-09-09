package main

import (
	"github.com/gin-contrib/pprof"
	"github.com/robfig/cron/v3"
	"log"
	"os"
	"time"
	"web-api/conf"
	"web-api/model"
	"web-api/server"
	"web-api/task"
	"web-api/task/websocket"
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
		go task.ProcessSabaSettle()
		go func() {
			for {
				websocket.SetupWebsocket()
				<-websocket.Websocket.Ended
				time.Sleep(1 * time.Second)
			}
		}()

		c := cron.New(cron.WithSeconds())
		c.Start()
		select {}
	} else {
		r := server.NewRouter()
		pprof.Register(r)
		r.Run(":" + os.Getenv("PORT"))
	}

	defer model.IPDB.Close()
}
