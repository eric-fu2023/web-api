package task

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"context"
	"encoding/json"
	"github.com/eclipse/paho.golang/paho"
	"strings"
	"web-api/model"
	"web-api/serializer"
	"web-api/util"
)

type Subscribe struct {
	util.BaseMQTTMsg
	Topic string `json:"topic"`
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

func SendPrivateChatHistory() {
	go func() {
		for {
			select {
			case msg, ok := <-util.TopicChannels["/subscribed"]:
				if !ok { // channel closed
					return
				}
				var v Subscribe
				if e := json.Unmarshal(msg, &v); e == nil {
					var userRef string
					if strings.Contains(v.Topic, "/cs/user") {
						userRef = v.Username
					} else if strings.Contains(v.Topic, "/cs/guest") {
						userRef = v.ClientId
					}
					if userRef != "" {
						go func(userRef string, v Subscribe) {
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
