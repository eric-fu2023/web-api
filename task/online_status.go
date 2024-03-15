package task

import (
	"context"
	"encoding/json"
	"github.com/eclipse/paho.golang/paho"
	"web-api/cache"
	"web-api/util"
)

type ConnectDisconnect struct {
	util.BaseMQTTMsg
}

func init() {
	subscription := paho.SubscribeOptions{Topic: "$SYS/brokers/#", QoS: 1}
	var exists bool
	for _, s := range util.Subscriptions {
		if s.Topic == subscription.Topic {
			exists = true
			break
		}
	}
	if !exists {
		util.Subscriptions = append(util.Subscriptions, subscription)
	}
	topics := []string{"/connected", "/disconnected"}
	for _, topic := range topics {
		var ex bool
		for t := range util.TopicChannels {
			if t == topic {
				ex = true
				break
			}
		}
		if !ex {
			util.TopicChannels[topic] = make(chan []byte, 100)
		}
	}
}

func UpdateOnlineStatus() {
	go func() {
		for {
			select {
			case msg, ok := <-util.TopicChannels["/connected"]:
				if !ok { // channel closed
					return
				}
				var v ConnectDisconnect
				if e := json.Unmarshal(msg, &v); e == nil {
					if v.Username == "admin" {
						// do nothing
					} else if v.Username == "test" {
						cache.RedisSessionClient.SAdd(context.TODO(), "online_guests", v.ClientId)
					} else {
						cache.RedisSessionClient.SAdd(context.TODO(), "online_users", v.Username)
					}
				}
			}
		}
	}()
	go func() {
		for {
			select {
			case msg, ok := <-util.TopicChannels["/disconnected"]:
				if !ok { // channel closed
					return
				}
				var v ConnectDisconnect
				if e := json.Unmarshal(msg, &v); e == nil {
					if v.Username == "admin" {
						// do nothing
					} else if v.Username == "test" {
						cache.RedisSessionClient.SRem(context.TODO(), "online_guests", v.ClientId)
					} else {
						cache.RedisSessionClient.SRem(context.TODO(), "online_users", v.Username)
					}
				}
			}
		}
	}()
}
