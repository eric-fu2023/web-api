package task

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"context"
	"encoding/json"
	"fmt"
	"github.com/eclipse/paho.golang/paho"
	"os"
	"strings"
	"web-api/cache"
	"web-api/model"
	"web-api/serializer"
	"web-api/util"
)

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
	topics := []string{"/subscribed"}
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

func UpdateSubscribed() {
	go func() {
		for {
			select {
			case msg, ok := <-util.TopicChannels["/subscribed"]:
				if !ok { // channel closed
					return
				}
				var v SubscribeUnsubscribe
				if e := json.Unmarshal(msg, &v); e == nil {
					var userRef string
					if v.Username != "admin" {
						if strings.Contains(v.Topic, "/cs/user") {
							if v.Username != os.Getenv("MQTT_GUEST_USERNAME") {
								userRef = v.Username
								go cache.RedisSessionClient.SAdd(context.TODO(), fmt.Sprintf(RedisKeySubscribedUsers, "cs"), userRef)
							}
						} else if strings.Contains(v.Topic, "/cs/guest") {
							userRef = v.ClientId
							go cache.RedisSessionClient.SAdd(context.TODO(), fmt.Sprintf(RedisKeySubscribedGuests, "cs"), userRef)
						}
					}
					if userRef != "" {
						go func(userRef string, v SubscribeUnsubscribe) {
							var messages []ploutos.PrivateMessage
							err := model.DB.Model(ploutos.PrivateMessage{}).Where(`user_ref = ? OR user_ref = '0'`, userRef).Order(`created_at DESC`).Limit(10).Find(&messages).Error
							if err != nil {
								util.Log().Error("private message history error", err)
								return
							}
							for i := len(messages) - 1; i >= 0; i-- {
								message := serializer.BuildPrivateMessage(messages[i])
								if j, e := json.Marshal(message); e == nil {
									pb := &paho.Publish{
										Topic:   v.Topic,
										QoS:     byte(1),
										Payload: j,
									}
									_, err = util.MQTTClient.Publish(context.Background(), pb)
									if err != nil {
										util.Log().Error("private message history error", err)
									}
								}
							}
						}(userRef, v)
					}
				}
			}
		}
	}()
}
