package util

import (
	"context"
	"fmt"
	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	"github.com/google/uuid"
	"net/url"
	"os"
	"strings"
	"time"
)

type BaseMQTTMsg struct {
	Username string `json:"username"`
	ClientId string `json:"clientid"`
}

type Subscribe struct {
	BaseMQTTMsg
	Topic string `json:"topic"`
}

var (
	MQTTClient    *autopaho.ConnectionManager
	Subscriptions []paho.SubscribeOptions
	TopicChannels = make(map[string]chan []byte)
)

func InitMQTT() {
	ctx := context.TODO()
	u, err := url.Parse(fmt.Sprintf("mqtts://%s", os.Getenv("MQTT_ADDRESS")))
	if err != nil {
		panic(err)
	}
	clientId := uuid.NewString()
	cfg := autopaho.ClientConfig{
		BrokerUrls:        []*url.URL{u},
		KeepAlive:         20,
		ConnectTimeout:    3 * time.Second,
		ConnectRetryDelay: 3 * time.Second,
		OnConnectionUp: func(cm *autopaho.ConnectionManager, connAck *paho.Connack) {
			fmt.Println("mqtt connection up")
			if _, err := cm.Subscribe(context.Background(), &paho.Subscribe{
				Subscriptions: Subscriptions,
			}); err != nil {
				fmt.Printf("failed to subscribe (%s). This is likely to mean no messages will be received.", err)
			}
			fmt.Println("mqtt subscription made")
		},
		OnConnectError: func(err error) { fmt.Printf("mqtt error whilst attempting connection: %s\n", err) },
		ClientConfig: paho.ClientConfig{
			ClientID:      clientId,
			OnClientError: func(err error) { fmt.Printf("mqtt connection error: %s\n", err) },
			OnServerDisconnect: func(d *paho.Disconnect) {
				if d.Properties != nil {
					fmt.Printf("mqtt server error: %s\n", d.Properties.ReasonString)
				} else {
					fmt.Printf("mqtt server error; code: %d\n", d.ReasonCode)
				}
			},
			OnPublishReceived: []func(paho.PublishReceived) (bool, error){
				func(pr paho.PublishReceived) (bool, error) {
					for topic, channel := range TopicChannels {
						if strings.Contains(pr.Packet.Topic, topic) {
							channel <- pr.Packet.Payload
							break
						}
					}
					return true, nil
				},
			},
		},
	}
	cfg.SetUsernamePassword(os.Getenv("MQTT_USERNAME"), []byte(os.Getenv("MQTT_PASSWORD")))
	c, err := autopaho.NewConnection(ctx, cfg)
	if err != nil {
		panic(err)
	}
	if err = c.AwaitConnection(ctx); err != nil {
		panic(err)
	}
	MQTTClient = c
}
