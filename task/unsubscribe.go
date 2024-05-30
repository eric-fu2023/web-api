package task

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/eclipse/paho.golang/paho"
	"os"
	"strings"
	"web-api/cache"
	"web-api/util"
)

func UpdateUnsubscribed() {
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
	topics := []string{"/unsubscribed"}
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

	go func() {
		for {
			select {
			case msg, ok := <-util.TopicChannels["/unsubscribed"]:
				if !ok { // channel closed
					return
				}
				var v SubscribeUnsubscribe
				if e := json.Unmarshal(msg, &v); e == nil {
					if v.Username != "admin" {
						if strings.Contains(v.Topic, "/cs/user") {
							if v.Username != os.Getenv("MQTT_GUEST_USERNAME") {
								go cache.RedisSessionClient.SRem(context.TODO(), fmt.Sprintf(RedisKeySubscribedUsers, "cs"), v.Username)
							}
						} else if strings.Contains(v.Topic, "/cs/guest") {
							go cache.RedisSessionClient.SRem(context.TODO(), fmt.Sprintf(RedisKeySubscribedGuests, "cs"), v.ClientId)
						}
					}
				}
			}
		}
	}()
}
